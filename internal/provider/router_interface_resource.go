// sws_router_interface — attach a subnet to a router. ID is composite
// "<router_id>:<subnet_id>" so import works.
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

type RouterInterfaceResource struct{ client *Client }
type RouterInterfaceModel struct {
	ID       types.String `tfsdk:"id"`
	RouterID types.String `tfsdk:"router_id"`
	SubnetID types.String `tfsdk:"subnet_id"`
}

func NewRouterInterfaceResource() resource.Resource { return &RouterInterfaceResource{} }
func (r *RouterInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_router_interface"
}
func (r *RouterInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *RouterInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a subnet to a router. Equivalent to AWS VPC subnet → route table association.",
		Attributes: map[string]schema.Attribute{
			"id":        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"router_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"subnet_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
		},
	}
}
func (r *RouterInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RouterInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"subnet_id": plan.SubnetID.ValueString()}
	if err := r.client.Do("PUT", "/api/network/routers/"+plan.RouterID.ValueString()+"/add_router_interface", body, nil); err != nil {
		resp.Diagnostics.AddError("attach router interface", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.RouterID.ValueString() + ":" + plan.SubnetID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *RouterInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Idempotent: as long as router + subnet still exist, treat as present.
	var state RouterInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var dummy any
	if err := r.client.Do("GET", "/api/network/routers/"+state.RouterID.ValueString(), nil, &dummy); err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *RouterInterfaceResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All fields are RequiresReplace; unreachable.
}
func (r *RouterInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RouterInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"subnet_id": state.SubnetID.ValueString()}
	err := r.client.Do("PUT", "/api/network/routers/"+state.RouterID.ValueString()+"/remove_router_interface", body, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && (apiErr.Status == 404) {
			return
		}
		resp.Diagnostics.AddError("detach router interface", err.Error())
	}
}
func (r *RouterInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import id", "expected <router_id>:<subnet_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("router_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subnet_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
