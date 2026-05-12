// sws_serverless_container — Zun container.
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

type ServerlessContainerResource struct{ client *Client }
type ServerlessContainerModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Image     types.String `tfsdk:"image"`
	NetworkID types.String `tfsdk:"network_id"`
	Status    types.String `tfsdk:"status"`
	Address   types.String `tfsdk:"address"`
}

func NewServerlessContainerResource() resource.Resource { return &ServerlessContainerResource{} }
func (r *ServerlessContainerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverless_container"
}
func (r *ServerlessContainerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ServerlessContainerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Zun (serverless) container.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":       schema.StringAttribute{Required: true, PlanModifiers: replace},
			"image":      schema.StringAttribute{Required: true, Description: "Docker image, e.g. nginx:alpine", PlanModifiers: replace},
			"network_id": schema.StringAttribute{Optional: true, PlanModifiers: append(keep, replace...)},
			"status":     schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"address":    schema.StringAttribute{Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *ServerlessContainerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServerlessContainerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString(), "image": plan.Image.ValueString()}
	if !plan.NetworkID.IsNull() && !plan.NetworkID.IsUnknown() && plan.NetworkID.ValueString() != "" {
		body["network_id"] = plan.NetworkID.ValueString()
	}
	var got struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Address string `json:"address"`
	}
	if err := r.client.Do("POST", "/api/serverless/containers", body, &got); err != nil {
		resp.Diagnostics.AddError("create container", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.Status = types.StringValue(got.Status)
	plan.Address = types.StringValue(got.Address)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *ServerlessContainerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServerlessContainerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		Status  string `json:"status"`
		Address string `json:"address"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/serverless/containers/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read container", err.Error())
		return
	}
	state.Status = types.StringValue(got.Status)
	state.Address = types.StringValue(got.Address)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *ServerlessContainerResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}
func (r *ServerlessContainerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServerlessContainerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/serverless/containers/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete container", err.Error())
	}
}
func (r *ServerlessContainerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
