// sws_volume — block storage volume (Cinder/RBD on Savannaa).
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VolumeResource struct{ client *Client }
type VolumeModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Size             types.Int64  `tfsdk:"size"`
	Description      types.String `tfsdk:"description"`
	VolumeType       types.String `tfsdk:"volume_type"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	Status           types.String `tfsdk:"status"`
}

func NewVolumeResource() resource.Resource { return &VolumeResource{} }
func (r *VolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}
func (r *VolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *VolumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replaceStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	replaceInt := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Cinder block-storage volume backed by Ceph RBD.",
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":              schema.StringAttribute{Required: true},
			"size":              schema.Int64Attribute{Required: true, Description: "Size in GiB", PlanModifiers: replaceInt},
			"description":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: keep},
			"volume_type":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: append(keep, replaceStr...)},
			"availability_zone": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: append(keep, replaceStr...)},
			"status":            schema.StringAttribute{Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VolumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"name": plan.Name.ValueString(),
		"size": plan.Size.ValueInt64(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.VolumeType.IsNull() && !plan.VolumeType.IsUnknown() && plan.VolumeType.ValueString() != "" {
		body["volume_type"] = plan.VolumeType.ValueString()
	}
	if !plan.AvailabilityZone.IsNull() && !plan.AvailabilityZone.IsUnknown() && plan.AvailabilityZone.ValueString() != "" {
		body["availability_zone"] = plan.AvailabilityZone.ValueString()
	}
	var got struct {
		ID               string `json:"id"`
		Status           string `json:"status"`
		VolumeType       string `json:"volume_type"`
		AvailabilityZone string `json:"availability_zone"`
		Description      string `json:"description"`
	}
	if err := r.client.Do("POST", "/api/storage/volumes", body, &got); err != nil {
		resp.Diagnostics.AddError("create volume", err.Error())
		return
	}
	// Poll for `available` (or `error`) — Cinder returns immediately with `creating`.
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		var v struct {
			Status string `json:"status"`
		}
		if err := r.client.Do("GET", "/api/storage/volumes/"+got.ID, nil, &v); err == nil {
			if v.Status == "available" || v.Status == "error" {
				got.Status = v.Status
				break
			}
		}
		time.Sleep(3 * time.Second)
	}
	plan.ID = types.StringValue(got.ID)
	plan.Status = types.StringValue(got.Status)
	plan.VolumeType = types.StringValue(got.VolumeType)
	plan.AvailabilityZone = types.StringValue(got.AvailabilityZone)
	plan.Description = types.StringValue(got.Description)
	if got.Status == "error" {
		resp.Diagnostics.AddError("create volume", "volume entered error state")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VolumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		Size             int64  `json:"size"`
		Status           string `json:"status"`
		VolumeType       string `json:"volume_type"`
		AvailabilityZone string `json:"availability_zone"`
		Description      string `json:"description"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/storage/volumes/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read volume", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.Status = types.StringValue(got.Status)
	state.VolumeType = types.StringValue(got.VolumeType)
	state.AvailabilityZone = types.StringValue(got.AvailabilityZone)
	state.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VolumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Description.IsNull() {
		body["description"] = plan.Description.ValueString()
	}
	_ = r.client.Do("PUT", "/api/storage/volumes/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VolumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/storage/volumes/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete volume", err.Error())
	}
}
func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
