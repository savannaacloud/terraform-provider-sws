// sws_tag — A free-form tag attached to project resources.
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

type TagResource struct{ client *Client }
type TagModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Config types.String `tfsdk:"config"`
	Status types.String `tfsdk:"status"`
}

func NewTagResource() resource.Resource { return &TagResource{} }
func (r *TagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}
func (r *TagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	c, ok := req.ProviderData.(*Client)
	if !ok { resp.Diagnostics.AddError("internal", fmt.Sprintf("expected *Client, got %T", req.ProviderData)); return }
	r.client = c
}
func (r *TagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A free-form tag attached to project resources.",
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":   schema.StringAttribute{Required: true},
			"config": schema.StringAttribute{Optional: true, Computed: true, Description: "JSON-encoded service-specific configuration. See backend docs.", PlanModifiers: keep},
			"status": schema.StringAttribute{Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() && plan.Config.ValueString() != "" {
		body["config"] = plan.Config.ValueString()
	}
	var got struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Config string `json:"config"`
	}
	if err := r.client.Do("POST", "/api/tags", body, &got); err != nil {
		resp.Diagnostics.AddError("create tag", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.Status = types.StringValue(got.Status)
	plan.Config = types.StringValue(got.Config)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	var got struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Config string `json:"config"`
	}
	err := r.client.Do("GET", "/api/tags/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read tag", err.Error())
		return
	}
	if got.Name != "" { state.Name = types.StringValue(got.Name) }
	state.Status = types.StringValue(got.Status)
	if got.Config != "" { state.Config = types.StringValue(got.Config) }
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Config.IsNull() { body["config"] = plan.Config.ValueString() }
	_ = r.client.Do("PUT", "/api/tags/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	err := r.client.Do("DELETE", "/api/tags/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 { return }
		resp.Diagnostics.AddError("delete tag", err.Error())
	}
}
func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
