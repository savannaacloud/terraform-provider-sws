// sws_managed_database — Trove DBaaS instance (mysql, postgres, etc.).
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ManagedDatabaseResource struct{ client *Client }
type ManagedDatabaseModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Datastore  types.String `tfsdk:"datastore"`
	Version    types.String `tfsdk:"version"`
	FlavorID   types.String `tfsdk:"flavor_id"`
	Size       types.Int64  `tfsdk:"size"`
	NetworkID  types.String `tfsdk:"network_id"`
	RootEnabled types.Bool  `tfsdk:"root_enabled"`
	Status     types.String `tfsdk:"status"`
	Address    types.String `tfsdk:"address"`
}

func NewManagedDatabaseResource() resource.Resource { return &ManagedDatabaseResource{} }
func (r *ManagedDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_database"
}
func (r *ManagedDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ManagedDatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A managed database instance (Trove). datastore = mysql / postgres / mariadb. Polls for ACTIVE (3-6 min).",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":         schema.StringAttribute{Required: true, PlanModifiers: replace},
			"datastore":    schema.StringAttribute{Required: true, Description: "mysql, postgres, mariadb", PlanModifiers: replace},
			"version":      schema.StringAttribute{Required: true, Description: "e.g. 8.0 for mysql", PlanModifiers: replace},
			"flavor_id":    schema.StringAttribute{Required: true},
			"size":         schema.Int64Attribute{Required: true, Description: "Volume size in GiB"},
			"network_id":   schema.StringAttribute{Required: true, PlanModifiers: replace},
			"root_enabled": schema.BoolAttribute{Optional: true, Computed: true},
			"status":       schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"address":      schema.StringAttribute{Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *ManagedDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ManagedDatabaseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"name":       plan.Name.ValueString(),
		"datastore":  plan.Datastore.ValueString(),
		"version":    plan.Version.ValueString(),
		"flavor_id":  plan.FlavorID.ValueString(),
		"size":       plan.Size.ValueInt64(),
		"network_id": plan.NetworkID.ValueString(),
	}
	var got struct {
		ID string `json:"id"`
	}
	if err := r.client.Do("POST", "/api/database/instances", body, &got); err != nil {
		resp.Diagnostics.AddError("create managed db", err.Error())
		return
	}
	// Poll for ACTIVE.
	deadline := time.Now().Add(10 * time.Minute)
	var status, addr string
	for time.Now().Before(deadline) {
		var c struct {
			Status  string `json:"status"`
			Address string `json:"address"`
			IP      string `json:"ip"`
		}
		if err := r.client.Do("GET", "/api/database/instances/"+got.ID, nil, &c); err == nil {
			status = c.Status
			addr = c.Address
			if addr == "" {
				addr = c.IP
			}
			if status == "ACTIVE" || status == "ERROR" {
				break
			}
		}
		time.Sleep(15 * time.Second)
	}
	plan.ID = types.StringValue(got.ID)
	plan.Status = types.StringValue(status)
	plan.Address = types.StringValue(addr)
	rootEnabled := false
	if !plan.RootEnabled.IsNull() && !plan.RootEnabled.IsUnknown() {
		rootEnabled = plan.RootEnabled.ValueBool()
	}
	if rootEnabled && status == "ACTIVE" {
		_ = r.client.Do("POST", "/api/database/instances/"+got.ID+"/root", nil, nil)
	}
	// Computed attribute must have a known value after apply.
	plan.RootEnabled = types.BoolValue(rootEnabled)
	if status == "ERROR" {
		resp.Diagnostics.AddError("managed db create failed", "instance entered ERROR state")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *ManagedDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ManagedDatabaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Address string `json:"address"`
		IP      string `json:"ip"`
		Size    int64  `json:"size"`
	}
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	err := r.client.Do("GET", "/api/database/instances/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read managed db", err.Error())
		return
	}
	state.Status = types.StringValue(got.Status)
	if got.Address != "" {
		state.Address = types.StringValue(got.Address)
	} else if got.IP != "" {
		state.Address = types.StringValue(got.IP)
	}
	state.Size = types.Int64Value(got.Size)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *ManagedDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ManagedDatabaseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"flavor_id": plan.FlavorID.ValueString(),
		"size":      plan.Size.ValueInt64(),
	}
	_ = r.client.Do("PUT", "/api/database/instances/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *ManagedDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ManagedDatabaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/database/instances/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete managed db", err.Error())
	}
}
func (r *ManagedDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
