// sws_lb_member — backend member in an LB pool.
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LBMemberResource struct{ client *Client }
type LBMemberModel struct {
	ID           types.String `tfsdk:"id"`
	PoolID       types.String `tfsdk:"pool_id"`
	Address      types.String `tfsdk:"address"`
	ProtocolPort types.Int64  `tfsdk:"protocol_port"`
	SubnetID     types.String `tfsdk:"subnet_id"`
	Weight       types.Int64  `tfsdk:"weight"`
}

func NewLBMemberResource() resource.Resource { return &LBMemberResource{} }
func (r *LBMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_member"
}
func (r *LBMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *LBMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replaceStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	replaceInt := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A backend in an LB pool (typically an instance IP). Composite id <pool_id>:<member_id>.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"pool_id":       schema.StringAttribute{Required: true, PlanModifiers: replaceStr},
			"address":       schema.StringAttribute{Required: true, PlanModifiers: replaceStr},
			"protocol_port": schema.Int64Attribute{Required: true, PlanModifiers: replaceInt},
			"subnet_id":     schema.StringAttribute{Required: true, PlanModifiers: replaceStr},
			"weight":        schema.Int64Attribute{Optional: true, Computed: true, Description: "Member weight (default 1)"},
		},
	}
}
func (r *LBMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LBMemberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"address":       plan.Address.ValueString(),
		"protocol_port": plan.ProtocolPort.ValueInt64(),
		"subnet_id":     plan.SubnetID.ValueString(),
	}
	w := int64(1)
	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		w = plan.Weight.ValueInt64()
	}
	body["weight"] = w
	var got struct {
		ID string `json:"id"`
	}
	if err := r.client.Do("POST", "/api/orchestration/pools/"+plan.PoolID.ValueString()+"/members", body, &got); err != nil {
		resp.Diagnostics.AddError("create member", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.PoolID.ValueString() + ":" + got.ID)
	plan.Weight = types.Int64Value(w)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBMemberResource) memberID() string {
	return ""
}
func (r *LBMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LBMemberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	parts := strings.SplitN(state.ID.ValueString(), ":", 2)
	if len(parts) != 2 {
		resp.State.RemoveResource(ctx)
		return
	}
	var got struct {
		ID     string `json:"id"`
		Weight int64  `json:"weight"`
	}
	err := r.client.Do("GET", "/api/orchestration/pools/"+parts[0]+"/members/"+parts[1], nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read member", err.Error())
		return
	}
	state.Weight = types.Int64Value(got.Weight)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *LBMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LBMemberModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	parts := strings.SplitN(plan.ID.ValueString(), ":", 2)
	if len(parts) == 2 && !plan.Weight.IsNull() {
		body := map[string]any{"weight": plan.Weight.ValueInt64()}
		_ = r.client.Do("PUT", "/api/orchestration/pools/"+parts[0]+"/members/"+parts[1], body, nil)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LBMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LBMemberModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	parts := strings.SplitN(state.ID.ValueString(), ":", 2)
	if len(parts) != 2 {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/pools/"+parts[0]+"/members/"+parts[1], nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete member", err.Error())
	}
}
func (r *LBMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import id", "expected <pool_id>:<member_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pool_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
