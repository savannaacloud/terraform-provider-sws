// sws_lb_health_monitor — Octavia health monitor on a pool.
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

type LBHealthMonitorResource struct{ client *Client }
type LBHealthMonitorModel struct {
	ID         types.String `tfsdk:"id"`
	PoolID     types.String `tfsdk:"pool_id"`
	Type       types.String `tfsdk:"type"`
	Delay      types.Int64  `tfsdk:"delay"`
	Timeout    types.Int64  `tfsdk:"timeout"`
	MaxRetries types.Int64  `tfsdk:"max_retries"`
	URLPath    types.String `tfsdk:"url_path"`
}

func NewLBHealthMonitorResource() resource.Resource { return &LBHealthMonitorResource{} }
func (r *LBHealthMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_health_monitor"
}
func (r *LBHealthMonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *LBHealthMonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Octavia health monitor for a pool (one monitor per pool).",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"pool_id":     schema.StringAttribute{Required: true, PlanModifiers: replace},
			"type":        schema.StringAttribute{Required: true, Description: "TCP, HTTP, HTTPS, PING", PlanModifiers: replace},
			"delay":       schema.Int64Attribute{Required: true, Description: "Seconds between probes"},
			"timeout":     schema.Int64Attribute{Required: true, Description: "Probe timeout in seconds"},
			"max_retries": schema.Int64Attribute{Required: true, Description: "Failures before marking member down"},
			"url_path":    schema.StringAttribute{Optional: true, Computed: true, Description: "HTTP probe path (default /)", PlanModifiers: keep},
		},
	}
}
func (r *LBHealthMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LBHealthMonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"type":        plan.Type.ValueString(),
		"delay":       plan.Delay.ValueInt64(),
		"timeout":     plan.Timeout.ValueInt64(),
		"max_retries": plan.MaxRetries.ValueInt64(),
	}
	if !plan.URLPath.IsNull() && !plan.URLPath.IsUnknown() && plan.URLPath.ValueString() != "" {
		body["url_path"] = plan.URLPath.ValueString()
	}
	var got struct {
		ID      string `json:"id"`
		URLPath string `json:"url_path"`
	}
	if err := r.client.Do("POST", "/api/orchestration/pools/"+plan.PoolID.ValueString()+"/health-monitor", body, &got); err != nil {
		resp.Diagnostics.AddError("create health monitor", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	if got.URLPath != "" {
		plan.URLPath = types.StringValue(got.URLPath)
	} else {
		plan.URLPath = types.StringValue("/")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBHealthMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LBHealthMonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Delay      int64  `json:"delay"`
		Timeout    int64  `json:"timeout"`
		MaxRetries int64  `json:"max_retries"`
		URLPath    string `json:"url_path"`
	}
	err := r.client.Do("GET", "/api/orchestration/health-monitors/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read health monitor", err.Error())
		return
	}
	state.Delay = types.Int64Value(got.Delay)
	state.Timeout = types.Int64Value(got.Timeout)
	state.MaxRetries = types.Int64Value(got.MaxRetries)
	if got.URLPath != "" {
		state.URLPath = types.StringValue(got.URLPath)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *LBHealthMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LBHealthMonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"delay":       plan.Delay.ValueInt64(),
		"timeout":     plan.Timeout.ValueInt64(),
		"max_retries": plan.MaxRetries.ValueInt64(),
	}
	if !plan.URLPath.IsNull() {
		body["url_path"] = plan.URLPath.ValueString()
	}
	_ = r.client.Do("PUT", "/api/orchestration/health-monitors/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBHealthMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LBHealthMonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/health-monitors/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete health monitor", err.Error())
	}
}
func (r *LBHealthMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
