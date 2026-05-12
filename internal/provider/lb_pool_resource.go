// sws_lb_pool — Octavia backend pool.
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

type LBPoolResource struct{ client *Client }
type LBPoolModel struct {
	ID             types.String `tfsdk:"id"`
	LoadBalancerID types.String `tfsdk:"load_balancer_id"`
	Name           types.String `tfsdk:"name"`
	Protocol       types.String `tfsdk:"protocol"`
	LBAlgorithm    types.String `tfsdk:"lb_algorithm"`
}

func NewLBPoolResource() resource.Resource { return &LBPoolResource{} }
func (r *LBPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_pool"
}
func (r *LBPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *LBPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Octavia backend pool.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"load_balancer_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"name":             schema.StringAttribute{Required: true},
			"protocol":         schema.StringAttribute{Required: true, Description: "TCP, HTTP, HTTPS, PROXY", PlanModifiers: replace},
			"lb_algorithm":     schema.StringAttribute{Optional: true, Computed: true, Description: "ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP (default ROUND_ROBIN)", PlanModifiers: keep},
		},
	}
}
func (r *LBPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LBPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	algo := "ROUND_ROBIN"
	if !plan.LBAlgorithm.IsNull() && !plan.LBAlgorithm.IsUnknown() && plan.LBAlgorithm.ValueString() != "" {
		algo = plan.LBAlgorithm.ValueString()
	}
	body := map[string]any{
		"name":            plan.Name.ValueString(),
		"loadbalancer_id": plan.LoadBalancerID.ValueString(),
		"protocol":        plan.Protocol.ValueString(),
		"lb_algorithm":    algo,
	}
	var got struct {
		ID string `json:"id"`
	}
	if err := r.client.Do("POST", "/api/orchestration/load-balancers/"+plan.LoadBalancerID.ValueString()+"/pools", body, &got); err != nil {
		resp.Diagnostics.AddError("create pool", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.LBAlgorithm = types.StringValue(algo)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LBPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		LBAlgorithm string `json:"lb_algorithm"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/orchestration/pools/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read pool", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.LBAlgorithm = types.StringValue(got.LBAlgorithm)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *LBPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LBPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString(), "lb_algorithm": plan.LBAlgorithm.ValueString()}
	_ = r.client.Do("PUT", "/api/orchestration/pools/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LBPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/pools/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete pool", err.Error())
	}
}
func (r *LBPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
