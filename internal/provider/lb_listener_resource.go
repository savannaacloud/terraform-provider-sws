// sws_lb_listener — Octavia listener on a load balancer.
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

type LBListenerResource struct{ client *Client }
type LBListenerModel struct {
	ID             types.String `tfsdk:"id"`
	LoadBalancerID types.String `tfsdk:"load_balancer_id"`
	Name           types.String `tfsdk:"name"`
	Protocol       types.String `tfsdk:"protocol"`
	ProtocolPort   types.Int64  `tfsdk:"protocol_port"`
	DefaultPoolID  types.String `tfsdk:"default_pool_id"`
}

func NewLBListenerResource() resource.Resource { return &LBListenerResource{} }
func (r *LBListenerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_listener"
}
func (r *LBListenerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *LBListenerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Octavia listener on a load balancer (port + protocol).",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"load_balancer_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"name":             schema.StringAttribute{Required: true},
			"protocol":         schema.StringAttribute{Required: true, Description: "TCP, HTTP, HTTPS, TERMINATED_HTTPS", PlanModifiers: replace},
			"protocol_port":    schema.Int64Attribute{Required: true},
			"default_pool_id":  schema.StringAttribute{Optional: true},
		},
	}
}
func (r *LBListenerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LBListenerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"name":             plan.Name.ValueString(),
		"loadbalancer_id":  plan.LoadBalancerID.ValueString(),
		"protocol":         plan.Protocol.ValueString(),
		"protocol_port":    plan.ProtocolPort.ValueInt64(),
	}
	if !plan.DefaultPoolID.IsNull() && plan.DefaultPoolID.ValueString() != "" {
		body["default_pool_id"] = plan.DefaultPoolID.ValueString()
	}
	var got struct {
		ID string `json:"id"`
	}
	if err := r.client.Do("POST", "/api/orchestration/load-balancers/"+plan.LoadBalancerID.ValueString()+"/listeners", body, &got); err != nil {
		resp.Diagnostics.AddError("create listener", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBListenerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LBListenerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Protocol      string `json:"protocol"`
		ProtocolPort  int64  `json:"protocol_port"`
		DefaultPoolID string `json:"default_pool_id"`
	}
	err := r.client.Do("GET", "/api/orchestration/listeners/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read listener", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.ProtocolPort = types.Int64Value(got.ProtocolPort)
	if got.DefaultPoolID != "" {
		state.DefaultPoolID = types.StringValue(got.DefaultPoolID)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *LBListenerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LBListenerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString(), "protocol_port": plan.ProtocolPort.ValueInt64()}
	if !plan.DefaultPoolID.IsNull() {
		body["default_pool_id"] = plan.DefaultPoolID.ValueString()
	}
	_ = r.client.Do("PUT", "/api/orchestration/listeners/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LBListenerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/listeners/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete listener", err.Error())
	}
}
func (r *LBListenerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
