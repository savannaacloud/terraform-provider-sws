// sws_floating_ip — explicit public IP allocation. Pass instance_id to
// associate at create time; otherwise it just allocates the IP and you
// associate it later (or via sws_instance.public_ip=true).
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

type FloatingIPResource struct{ client *Client }
type FloatingIPModel struct {
	ID                types.String `tfsdk:"id"`
	Address           types.String `tfsdk:"address"`
	FloatingNetworkID types.String `tfsdk:"floating_network_id"`
	InstanceID        types.String `tfsdk:"instance_id"`
	Description       types.String `tfsdk:"description"`
}

func NewFloatingIPResource() resource.Resource { return &FloatingIPResource{} }
func (r *FloatingIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip"
}
func (r *FloatingIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *FloatingIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A public/floating IP. Optionally associated with an instance at create time.",
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"address":             schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"floating_network_id": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: keep},
			"instance_id":         schema.StringAttribute{Optional: true},
			"description":         schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *FloatingIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FloatingIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{}
	if !plan.FloatingNetworkID.IsNull() && !plan.FloatingNetworkID.IsUnknown() {
		body["floating_network_id"] = plan.FloatingNetworkID.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.InstanceID.IsNull() && plan.InstanceID.ValueString() != "" {
		body["server_id"] = plan.InstanceID.ValueString()
	}
	var got struct {
		ID                string `json:"id"`
		FloatingIPAddress string `json:"floating_ip_address"`
		FloatingNetworkID string `json:"floating_network_id"`
		Description       string `json:"description"`
	}
	if err := r.client.Do("POST", "/api/network/floating-ips", body, &got); err != nil {
		resp.Diagnostics.AddError("allocate floating ip", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.Address = types.StringValue(got.FloatingIPAddress)
	plan.FloatingNetworkID = types.StringValue(got.FloatingNetworkID)
	plan.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *FloatingIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FloatingIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID                string `json:"id"`
		FloatingIPAddress string `json:"floating_ip_address"`
		FloatingNetworkID string `json:"floating_network_id"`
		Description       string `json:"description"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/network/floating-ips/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read floating ip", err.Error())
		return
	}
	state.Address = types.StringValue(got.FloatingIPAddress)
	state.FloatingNetworkID = types.StringValue(got.FloatingNetworkID)
	state.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *FloatingIPResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All optional attributes are computed-from-create; FIP doesn't really
	// support meaningful Updates from this provider yet.
}
func (r *FloatingIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FloatingIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/network/floating-ips/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete floating ip", err.Error())
	}
}
func (r *FloatingIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
