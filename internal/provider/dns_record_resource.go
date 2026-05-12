// sws_dns_record — recordset in a public DNS zone.
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DNSRecordResource struct{ client *Client }
type DNSRecordModel struct {
	ID      types.String `tfsdk:"id"`
	ZoneID  types.String `tfsdk:"zone_id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	TTL     types.Int64  `tfsdk:"ttl"`
	Records types.List   `tfsdk:"records"`
}

func NewDNSRecordResource() resource.Resource { return &DNSRecordResource{} }
func (r *DNSRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}
func (r *DNSRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *DNSRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A recordset (A, AAAA, CNAME, MX, TXT, ...) in an sws_dns_zone. Composite id <zone_id>:<rrset_id>.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"zone_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"name":    schema.StringAttribute{Required: true, Description: "Fully qualified name ending in a dot, e.g. www.example.com.", PlanModifiers: replace},
			"type":    schema.StringAttribute{Required: true, Description: "A, AAAA, CNAME, MX, TXT, NS, SRV", PlanModifiers: replace},
			"ttl":     schema.Int64Attribute{Optional: true, Computed: true},
			"records": schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "Record values, e.g. [\"1.2.3.4\"]"},
		},
	}
}
func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DNSRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var recs []string
	plan.Records.ElementsAs(ctx, &recs, false)
	body := map[string]any{
		"name":    plan.Name.ValueString(),
		"type":    plan.Type.ValueString(),
		"records": recs,
	}
	if !plan.TTL.IsNull() && !plan.TTL.IsUnknown() {
		body["ttl"] = plan.TTL.ValueInt64()
	}
	var got struct {
		ID  string `json:"id"`
		TTL int64  `json:"ttl"`
	}
	if err := r.client.Do("POST", "/api/orchestration/dns/zones/"+plan.ZoneID.ValueString()+"/recordsets", body, &got); err != nil {
		resp.Diagnostics.AddError("create dns record", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.ZoneID.ValueString() + ":" + got.ID)
	plan.TTL = types.Int64Value(got.TTL)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DNSRecordModel
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
		Name    string   `json:"name"`
		Type    string   `json:"type"`
		TTL     int64    `json:"ttl"`
		Records []string `json:"records"`
	}
	err := r.client.Do("GET", "/api/orchestration/dns/zones/"+parts[0]+"/recordsets/"+parts[1], nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read dns record", err.Error())
		return
	}
	state.TTL = types.Int64Value(got.TTL)
	state.Records, _ = types.ListValueFrom(ctx, types.StringType, got.Records)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DNSRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	parts := strings.SplitN(plan.ID.ValueString(), ":", 2)
	if len(parts) == 2 {
		var recs []string
		plan.Records.ElementsAs(ctx, &recs, false)
		body := map[string]any{"records": recs}
		if !plan.TTL.IsNull() {
			body["ttl"] = plan.TTL.ValueInt64()
		}
		_ = r.client.Do("PUT", "/api/orchestration/dns/zones/"+parts[0]+"/recordsets/"+parts[1], body, nil)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DNSRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	parts := strings.SplitN(state.ID.ValueString(), ":", 2)
	if len(parts) != 2 {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/dns/zones/"+parts[0]+"/recordsets/"+parts[1], nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete dns record", err.Error())
	}
}
func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import id", "expected <zone_id>:<rrset_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
