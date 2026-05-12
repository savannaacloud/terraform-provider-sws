// sws_subnet — additional subnets on a network (for multi-subnet networks).
// If you only need one subnet per network, use sws_network's inline cidr.
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SubnetResource struct{ client *Client }
type SubnetModel struct {
	ID         types.String `tfsdk:"id"`
	NetworkID  types.String `tfsdk:"network_id"`
	Name       types.String `tfsdk:"name"`
	CIDR       types.String `tfsdk:"cidr"`
	IPVersion  types.Int64  `tfsdk:"ip_version"`
	GatewayIP  types.String `tfsdk:"gateway_ip"`
	EnableDHCP types.Bool   `tfsdk:"enable_dhcp"`
	DNSServers types.List   `tfsdk:"dns_nameservers"`
}

func NewSubnetResource() resource.Resource { return &SubnetResource{} }
func (r *SubnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}
func (r *SubnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SubnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Neutron subnet on an existing network.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"network_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"name":       schema.StringAttribute{Required: true},
			"cidr":       schema.StringAttribute{Required: true, PlanModifiers: replace},
			"ip_version": schema.Int64Attribute{Optional: true, Computed: true},
			"gateway_ip": schema.StringAttribute{Optional: true, Computed: true},
			"enable_dhcp": schema.BoolAttribute{Optional: true, Computed: true},
			"dns_nameservers": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Defaults to [\"1.1.1.1\", \"8.8.8.8\"] if omitted.",
			},
		},
	}
}
func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SubnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ipv := int64(4)
	if !plan.IPVersion.IsNull() && !plan.IPVersion.IsUnknown() {
		ipv = plan.IPVersion.ValueInt64()
	}
	dns := []string{"1.1.1.1", "8.8.8.8"}
	if !plan.DNSServers.IsNull() && !plan.DNSServers.IsUnknown() {
		dns = nil
		plan.DNSServers.ElementsAs(ctx, &dns, false)
	}
	body := map[string]any{
		"name":            plan.Name.ValueString(),
		"network_id":      plan.NetworkID.ValueString(),
		"cidr":            plan.CIDR.ValueString(),
		"ip_version":      ipv,
		"enable_dhcp":     plan.EnableDHCP.IsNull() || plan.EnableDHCP.ValueBool(),
		"dns_nameservers": dns,
	}
	if !plan.GatewayIP.IsNull() && !plan.GatewayIP.IsUnknown() && plan.GatewayIP.ValueString() != "" {
		body["gateway_ip"] = plan.GatewayIP.ValueString()
	}
	var got struct {
		ID         string   `json:"id"`
		GatewayIP  string   `json:"gateway_ip"`
		IPVersion  int64    `json:"ip_version"`
		EnableDHCP bool     `json:"is_dhcp_enabled"`
		DNS        []string `json:"dns_nameservers"`
	}
	if err := r.client.Do("POST", "/api/network/subnets", body, &got); err != nil {
		resp.Diagnostics.AddError("create subnet", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.GatewayIP = types.StringValue(got.GatewayIP)
	plan.IPVersion = types.Int64Value(got.IPVersion)
	plan.EnableDHCP = types.BoolValue(got.EnableDHCP)
	plan.DNSServers, _ = types.ListValueFrom(ctx, types.StringType, got.DNS)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SubnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		CIDR       string   `json:"cidr"`
		NetworkID  string   `json:"network_id"`
		GatewayIP  string   `json:"gateway_ip"`
		IPVersion  int64    `json:"ip_version"`
		EnableDHCP bool     `json:"is_dhcp_enabled"`
		DNS        []string `json:"dns_nameservers"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/network/subnets/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read subnet", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.GatewayIP = types.StringValue(got.GatewayIP)
	state.IPVersion = types.Int64Value(got.IPVersion)
	state.EnableDHCP = types.BoolValue(got.EnableDHCP)
	state.DNSServers, _ = types.ListValueFrom(ctx, types.StringType, got.DNS)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SubnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	_ = r.client.Do("PUT", "/api/network/subnets/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SubnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/network/subnets/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete subnet", err.Error())
	}
}
func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
