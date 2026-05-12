// sws_router — Neutron router. Created with an external gateway by default
// (so it's immediately usable for floating IPs). Use sws_router_interface to
// attach subnets.
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

type RouterResource struct{ client *Client }
type RouterModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	ExternalNetworkID types.String `tfsdk:"external_network_id"`
}

func NewRouterResource() resource.Resource { return &RouterResource{} }
func (r *RouterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_router"
}
func (r *RouterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *RouterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Neutron router with external gateway (auto-detects the project's external network if external_network_id is omitted).",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":        schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: keep},
			"external_network_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "External network UUID. If omitted, the project's default external network is auto-discovered.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RouterResource) discoverExternal() string {
	var net struct {
		ID string `json:"id"`
	}
	if err := r.client.Do("GET", "/api/network/external-network", nil, &net); err == nil {
		return net.ID
	}
	return ""
}

func (r *RouterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RouterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ext := ""
	if !plan.ExternalNetworkID.IsNull() && !plan.ExternalNetworkID.IsUnknown() {
		ext = plan.ExternalNetworkID.ValueString()
	}
	if ext == "" {
		ext = r.discoverExternal()
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if ext != "" {
		body["external_gateway_info"] = map[string]any{"network_id": ext}
	}
	var got struct {
		ID                  string         `json:"id"`
		Description         string         `json:"description"`
		ExternalGatewayInfo map[string]any `json:"external_gateway_info"`
	}
	if err := r.client.Do("POST", "/api/network/routers", body, &got); err != nil {
		resp.Diagnostics.AddError("create router", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.Description = types.StringValue(got.Description)
	if got.ExternalGatewayInfo != nil {
		if nid, ok := got.ExternalGatewayInfo["network_id"].(string); ok {
			plan.ExternalNetworkID = types.StringValue(nid)
		}
	} else {
		plan.ExternalNetworkID = types.StringValue(ext)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *RouterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RouterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID                  string         `json:"id"`
		Name                string         `json:"name"`
		Description         string         `json:"description"`
		ExternalGatewayInfo map[string]any `json:"external_gateway_info"`
	}
	err := r.client.Do("GET", "/api/network/routers/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read router", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.Description = types.StringValue(got.Description)
	if got.ExternalGatewayInfo != nil {
		if nid, ok := got.ExternalGatewayInfo["network_id"].(string); ok {
			state.ExternalNetworkID = types.StringValue(nid)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *RouterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RouterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Description.IsNull() {
		body["description"] = plan.Description.ValueString()
	}
	_ = r.client.Do("PUT", "/api/network/routers/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *RouterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RouterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/network/routers/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete router", err.Error())
	}
}
func (r *RouterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
