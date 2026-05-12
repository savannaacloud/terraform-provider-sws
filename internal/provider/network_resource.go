// sws_network — Neutron tenant network with optional inline subnet.
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetworkResource struct{ client *Client }

type NetworkModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CIDR        types.String `tfsdk:"cidr"`
	SubnetID    types.String `tfsdk:"subnet_id"`
}

func NewNetworkResource() resource.Resource { return &NetworkResource{} }

func (r *NetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("internal", fmt.Sprintf("expected *Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A tenant network with an inline IPv4 subnet (AWS VPC + default-subnet equivalent). " +
			"Set cidr=\"\" to skip subnet creation if you want to add subnets manually.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cidr": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IPv4 CIDR for the inline subnet. Default 10.0.0.0/24. Pass an empty string to skip subnet creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Computed:      true,
				Description:   "ID of the inline subnet, if one was created.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	var got struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := r.client.Do("POST", "/api/network/networks", body, &got); err != nil {
		resp.Diagnostics.AddError("create network", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.Description = types.StringValue(got.Description)

	// Auto-create a subnet so instances can actually boot on this network.
	// nova rejects a port-create on a network without a subnet
	// ("Network <id> requires a subnet in order to boot instances on"),
	// so a network without a subnet is essentially useless in v0.1 where
	// we don't yet have a separate sws_subnet resource.
	cidr := "10.0.0.0/24"
	if !plan.CIDR.IsNull() && !plan.CIDR.IsUnknown() {
		cidr = plan.CIDR.ValueString()
	}
	plan.CIDR = types.StringValue(cidr)
	plan.SubnetID = types.StringValue("")
	if cidr != "" {
		subBody := map[string]any{
			"name":             plan.Name.ValueString() + "-subnet",
			"network_id":       got.ID,
			"cidr":             cidr,
			"ip_version":       4,
			"enable_dhcp":      true,
			"dns_nameservers":  []string{"1.1.1.1", "8.8.8.8"},
		}
		var sub struct {
			ID string `json:"id"`
		}
		if err := r.client.Do("POST", "/api/network/subnets", subBody, &sub); err != nil {
			// Roll back: delete the network we just created so the user can retry cleanly.
			_ = r.client.Do("DELETE", "/api/network/networks/"+got.ID, nil, nil)
			resp.Diagnostics.AddError("create subnet", err.Error())
			return
		}
		plan.SubnetID = types.StringValue(sub.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	err := r.client.Do("GET", "/api/network/networks/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read network", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NetworkModel
	var state NetworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Description.IsNull() {
		body["description"] = plan.Description.ValueString()
	}
	if err := r.client.Do("PUT", "/api/network/networks/"+plan.ID.ValueString(), body, nil); err != nil {
		resp.Diagnostics.AddError("update network", err.Error())
		return
	}
	// CIDR/subnet are RequiresReplace; carry the state forward.
	plan.SubnetID = state.SubnetID
	plan.CIDR = state.CIDR
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Cascade-delete order matters. Neutron refuses to delete a network with
	// any ports attached, refuses to delete a subnet with a router interface
	// on it, and refuses to delete a router-interface port directly. Order:
	//   1. detach every router interface that touches this network's subnet
	//   2. delete the subnet (which cleans up DHCP ports owned by the subnet)
	//   3. delete the network
	//
	// Anyone running `auto_public_ip=true` on an sws_instance will have had
	// the python-backend attach the network to the project's HA router
	// without telling the provider, so a clean destroy MUST do this lookup.

	// 1. find router-interface ports on this network and detach them
	var ports []struct {
		ID          string `json:"id"`
		DeviceOwner string `json:"device_owner"`
		DeviceID    string `json:"device_id"`
		FixedIPs    []struct {
			SubnetID string `json:"subnet_id"`
		} `json:"fixed_ips"`
		NetworkID string `json:"network_id"`
	}
	if err := r.client.Do("GET", "/api/network/ports", nil, &ports); err == nil {
		for _, p := range ports {
			if p.NetworkID != state.ID.ValueString() {
				continue
			}
			// router_interface OR ha_router_replicated_interface OR centralized variants
			if len(p.DeviceOwner) > 14 && p.DeviceOwner[:15] == "network:ha_rout" ||
				len(p.DeviceOwner) > 8 && p.DeviceOwner[:9] == "network:r" {
				if len(p.FixedIPs) > 0 && p.DeviceID != "" {
					body := map[string]any{"subnet_id": p.FixedIPs[0].SubnetID}
					_ = r.client.Do("PUT", "/api/network/routers/"+p.DeviceID+"/remove_router_interface", body, nil)
				}
			}
		}
	}

	// 2. delete the inline subnet (kills its DHCP ports — asynchronously,
	// neutron-dhcp-agent takes a few seconds to clean them up)
	if !state.SubnetID.IsNull() && state.SubnetID.ValueString() != "" {
		_ = r.client.Do("DELETE", "/api/network/subnets/"+state.SubnetID.ValueString(), nil, nil)
	}

	// 3. delete the network — retry for up to 30s because DHCP port cleanup
	// is async. Neutron returns 409 "There are one or more ports still in
	// use on the network" until the dhcp-agent finishes; observed window is
	// 2-8s. Worth retrying instead of erroring out.
	var lastErr error
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		lastErr = r.client.Do("DELETE", "/api/network/networks/"+state.ID.ValueString(), nil, nil)
		if lastErr == nil {
			return
		}
		if apiErr, ok := lastErr.(*APIError); ok {
			if apiErr.Status == 404 {
				return
			}
			if apiErr.Status != 409 && apiErr.Status != 502 {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	resp.Diagnostics.AddError("delete network", lastErr.Error())
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
