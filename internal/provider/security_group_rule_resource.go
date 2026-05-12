// sws_security_group_rule — one ingress/egress rule on a security group.
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityGroupRuleResource struct{ client *Client }
type SecurityGroupRuleModel struct {
	ID              types.String `tfsdk:"id"`
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	Direction       types.String `tfsdk:"direction"`
	Protocol        types.String `tfsdk:"protocol"`
	EtherType       types.String `tfsdk:"ethertype"`
	PortRangeMin    types.Int64  `tfsdk:"port_range_min"`
	PortRangeMax    types.Int64  `tfsdk:"port_range_max"`
	RemoteIPPrefix  types.String `tfsdk:"remote_ip_prefix"`
	Description     types.String `tfsdk:"description"`
}

func NewSecurityGroupRuleResource() resource.Resource { return &SecurityGroupRuleResource{} }
func (r *SecurityGroupRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_rule"
}
func (r *SecurityGroupRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SecurityGroupRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	replaceInt := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "An ingress/egress rule on a security group. Every attribute is RequiresReplace — rules are immutable; updates delete + recreate.",
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"security_group_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"direction":         schema.StringAttribute{Required: true, Description: "ingress or egress", PlanModifiers: replace},
			"protocol":          schema.StringAttribute{Optional: true, Computed: true, Description: "tcp, udp, icmp, or null for any", PlanModifiers: append([]planmodifier.String{stringplanmodifier.UseStateForUnknown()}, replace...)},
			"ethertype":         schema.StringAttribute{Optional: true, Computed: true, Description: "IPv4 or IPv6 (default IPv4)", PlanModifiers: append([]planmodifier.String{stringplanmodifier.UseStateForUnknown()}, replace...)},
			"port_range_min":    schema.Int64Attribute{Optional: true, PlanModifiers: replaceInt},
			"port_range_max":    schema.Int64Attribute{Optional: true, PlanModifiers: replaceInt},
			"remote_ip_prefix":  schema.StringAttribute{Optional: true, Description: "e.g. 0.0.0.0/0", PlanModifiers: replace},
			"description":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *SecurityGroupRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityGroupRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	et := "IPv4"
	if !plan.EtherType.IsNull() && !plan.EtherType.IsUnknown() && plan.EtherType.ValueString() != "" {
		et = plan.EtherType.ValueString()
	}
	body := map[string]any{
		"security_group_id": plan.SecurityGroupID.ValueString(),
		"direction":         plan.Direction.ValueString(),
		"ethertype":         et,
	}
	if !plan.Protocol.IsNull() && !plan.Protocol.IsUnknown() && plan.Protocol.ValueString() != "" {
		body["protocol"] = plan.Protocol.ValueString()
	}
	if !plan.PortRangeMin.IsNull() && !plan.PortRangeMin.IsUnknown() {
		body["port_range_min"] = plan.PortRangeMin.ValueInt64()
	}
	if !plan.PortRangeMax.IsNull() && !plan.PortRangeMax.IsUnknown() {
		body["port_range_max"] = plan.PortRangeMax.ValueInt64()
	}
	if !plan.RemoteIPPrefix.IsNull() && !plan.RemoteIPPrefix.IsUnknown() {
		body["remote_ip_prefix"] = plan.RemoteIPPrefix.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	var got struct {
		ID          string `json:"id"`
		Protocol    string `json:"protocol"`
		EtherType   string `json:"ethertype"`
		Description string `json:"description"`
	}
	if err := r.client.Do("POST", "/api/network/security-group-rules", body, &got); err != nil {
		resp.Diagnostics.AddError("create sg rule", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	if got.Protocol != "" {
		plan.Protocol = types.StringValue(got.Protocol)
	}
	plan.EtherType = types.StringValue(got.EtherType)
	plan.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *SecurityGroupRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityGroupRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID string `json:"id"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/network/security-group-rules/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		// Some backends return 400/502 wrapping 404; treat as gone if id missing.
		if apiErr, ok := err.(*APIError); ok && apiErr.Status >= 400 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read sg rule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *SecurityGroupRuleResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All RequiresReplace; unreachable.
}
func (r *SecurityGroupRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecurityGroupRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/network/security-group-rules/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete sg rule", err.Error())
	}
}
func (r *SecurityGroupRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
