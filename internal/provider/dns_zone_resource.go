// sws_dns_zone — public DNS zone managed by Designate.
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

type DNSZoneResource struct{ client *Client }
type DNSZoneModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Email       types.String `tfsdk:"email"`
	TTL         types.Int64  `tfsdk:"ttl"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Status      types.String `tfsdk:"status"`
}

func NewDNSZoneResource() resource.Resource { return &DNSZoneResource{} }
func (r *DNSZoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}
func (r *DNSZoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *DNSZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A public DNS zone (Designate). Use sws_dns_record to add records.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":        schema.StringAttribute{Required: true, Description: "Domain ending in a dot, e.g. example.com.", PlanModifiers: replace},
			"email":       schema.StringAttribute{Required: true, Description: "Zone admin email (replaces @ with . in SOA)"},
			"ttl":         schema.Int64Attribute{Optional: true, Computed: true, Description: "SOA TTL in seconds (default 3600)"},
			"description": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: keep},
			"type":        schema.StringAttribute{Optional: true, Computed: true, Description: "PRIMARY (default) or SECONDARY", PlanModifiers: append(keep, replace...)},
			"status":      schema.StringAttribute{Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *DNSZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DNSZoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"name":  plan.Name.ValueString(),
		"email": plan.Email.ValueString(),
	}
	if !plan.TTL.IsNull() && !plan.TTL.IsUnknown() {
		body["ttl"] = plan.TTL.ValueInt64()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() && plan.Type.ValueString() != "" {
		body["type"] = plan.Type.ValueString()
	}
	var got struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		Type        string `json:"type"`
		TTL         int64  `json:"ttl"`
		Description string `json:"description"`
	}
	if err := r.client.Do("POST", "/api/orchestration/dns/zones", body, &got); err != nil {
		resp.Diagnostics.AddError("create dns zone", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.Status = types.StringValue(got.Status)
	plan.Type = types.StringValue(got.Type)
	plan.TTL = types.Int64Value(got.TTL)
	plan.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *DNSZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DNSZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		TTL         int64  `json:"ttl"`
		Status      string `json:"status"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/orchestration/dns/zones/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read dns zone", err.Error())
		return
	}
	state.Email = types.StringValue(got.Email)
	state.TTL = types.Int64Value(got.TTL)
	state.Status = types.StringValue(got.Status)
	state.Type = types.StringValue(got.Type)
	state.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *DNSZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DNSZoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"email": plan.Email.ValueString()}
	if !plan.TTL.IsNull() {
		body["ttl"] = plan.TTL.ValueInt64()
	}
	if !plan.Description.IsNull() {
		body["description"] = plan.Description.ValueString()
	}
	_ = r.client.Do("PUT", "/api/orchestration/dns/zones/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *DNSZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DNSZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/dns/zones/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete dns zone", err.Error())
	}
}
func (r *DNSZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
