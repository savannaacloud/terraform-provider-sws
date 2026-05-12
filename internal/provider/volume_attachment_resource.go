// sws_volume_attachment — attach a volume to an instance.
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

type VolumeAttachmentResource struct{ client *Client }
type VolumeAttachmentModel struct {
	ID         types.String `tfsdk:"id"`
	InstanceID types.String `tfsdk:"instance_id"`
	VolumeID   types.String `tfsdk:"volume_id"`
	Device     types.String `tfsdk:"device"`
}

func NewVolumeAttachmentResource() resource.Resource { return &VolumeAttachmentResource{} }
func (r *VolumeAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_attachment"
}
func (r *VolumeAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *VolumeAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches an sws_volume to an sws_instance. Composite ID is <instance_id>:<volume_id>.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"instance_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"volume_id":   schema.StringAttribute{Required: true, PlanModifiers: replace},
			"device":      schema.StringAttribute{Optional: true, Computed: true, Description: "e.g. /dev/vdb. Auto-assigned if omitted.", PlanModifiers: append(keep, replace...)},
		},
	}
}
func (r *VolumeAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VolumeAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"volume_id": plan.VolumeID.ValueString()}
	if !plan.Device.IsNull() && !plan.Device.IsUnknown() && plan.Device.ValueString() != "" {
		body["device"] = plan.Device.ValueString()
	}
	var got struct {
		Device string `json:"device"`
	}
	if err := r.client.Do("POST", "/api/compute/servers/"+plan.InstanceID.ValueString()+"/volumes", body, &got); err != nil {
		resp.Diagnostics.AddError("attach volume", err.Error())
		return
	}
	if got.Device != "" {
		plan.Device = types.StringValue(got.Device)
	}
	plan.ID = types.StringValue(plan.InstanceID.ValueString() + ":" + plan.VolumeID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *VolumeAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VolumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Best-effort: verify volume still attached to instance.
	var attachments []struct {
		VolumeID string `json:"volume_id"`
		Device   string `json:"device"`
	}
	if err := r.client.Do("GET", "/api/compute/servers/"+state.InstanceID.ValueString()+"/volumes", nil, &attachments); err == nil {
		found := false
		for _, a := range attachments {
			if a.VolumeID == state.VolumeID.ValueString() {
				if a.Device != "" {
					state.Device = types.StringValue(a.Device)
				}
				found = true
				break
			}
		}
		if !found {
			resp.State.RemoveResource(ctx)
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *VolumeAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All RequiresReplace; unreachable.
}
func (r *VolumeAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VolumeAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/compute/servers/"+state.InstanceID.ValueString()+"/volumes/"+state.VolumeID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("detach volume", err.Error())
	}
}
func (r *VolumeAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import id", "expected <instance_id>:<volume_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
