// sws_instance — compute server (virtual machine) on Savannaa.
//
// CRUD maps to:
//   create:  POST   /api/compute/servers
//   read:    GET    /api/compute/servers/{id}
//   delete:  DELETE /api/compute/servers/{id}
//
// Updates are not supported by the upstream API for the fields a user
// would change in HCL (flavor + image + network are immutable post-
// create). Changing any of those forces resource replacement; we mark
// them RequiresReplace in the schema below.
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

type InstanceResource struct {
	client *Client
}

type InstanceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Plan       types.String `tfsdk:"plan"`
	Image      types.String `tfsdk:"image"`
	NetworkID  types.String `tfsdk:"network_id"`
	KeypairID  types.String `tfsdk:"keypair"`
	PublicIP   types.Bool   `tfsdk:"public_ip"`
	IPAddress  types.String `tfsdk:"ip_address"`
	Status     types.String `tfsdk:"status"`
}

func NewInstanceResource() resource.Resource { return &InstanceResource{} }

func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	immutable := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Savannaa compute instance. Equivalent to AWS EC2 / OpenStack Nova server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server UUID assigned by Savannaa.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the instance.",
			},
			"plan": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Plan / flavor name, e.g. `m1.small`. Use the `sws_plan` data source to look up available plans.",
				PlanModifiers:       immutable,
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Image UUID. Use the `sws_image` data source to resolve by name.",
				PlanModifiers:       immutable,
			},
			"network_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network UUID. Omit to use the project's default network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"keypair": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of an `sws_keypair` to inject for SSH access. Cloud-init drops the public key in the default user's authorized_keys.",
				PlanModifiers:       immutable,
			},
			"public_ip": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "When true, allocate a public IP at create time. Default: true.",
			},
			"ip_address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Public IP if allocated, else the primary fixed IP.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server lifecycle state (BUILD, ACTIVE, ERROR, etc.).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InstanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]any{
		"name":           plan.Name.ValueString(),
		"plan":           plan.Plan.ValueString(),
		"image":          plan.Image.ValueString(),
		"auto_public_ip": !plan.PublicIP.IsNull() && plan.PublicIP.ValueBool() || plan.PublicIP.IsNull(),
	}
	if !plan.NetworkID.IsNull() && plan.NetworkID.ValueString() != "" {
		body["networks"] = []map[string]string{{"uuid": plan.NetworkID.ValueString()}}
	}
	if !plan.KeypairID.IsNull() && plan.KeypairID.ValueString() != "" {
		body["key_name"] = plan.KeypairID.ValueString()
	}

	var created struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Status    string `json:"status"`
		IPAddress string `json:"ip_address"`
	}
	if err := r.client.Do("POST", "/api/compute/servers", body, &created); err != nil {
		resp.Diagnostics.AddError("create instance", err.Error())
		return
	}

	// Poll for ACTIVE (or ERROR) — Nova create returns immediately with BUILD.
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		var got struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			IPAddress string `json:"ip_address"`
			PublicIP  string `json:"public_ip"`
		}
		if err := r.client.Do("GET", "/api/compute/servers/"+created.ID, nil, &got); err == nil {
			if got.Status == "ACTIVE" || got.Status == "ERROR" {
				plan.ID = types.StringValue(got.ID)
				plan.Status = types.StringValue(got.Status)
				ip := got.PublicIP
				if ip == "" {
					ip = got.IPAddress
				}
				plan.IPAddress = types.StringValue(ip)
				if got.Status == "ERROR" {
					resp.Diagnostics.AddError("instance create failed", "server entered ERROR state")
					return
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
				return
			}
		}
		time.Sleep(5 * time.Second)
	}
	resp.Diagnostics.AddError("instance create timed out", "server did not become ACTIVE within 5 min")
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InstanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Status    string `json:"status"`
		IPAddress string `json:"ip_address"`
		PublicIP  string `json:"public_ip"`
	}
	err := r.client.Do("GET", "/api/compute/servers/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read instance", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.Status = types.StringValue(got.Status)
	ip := got.PublicIP
	if ip == "" {
		ip = got.IPAddress
	}
	state.IPAddress = types.StringValue(ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *InstanceResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All updatable fields are RequiresReplace; this method is unreachable
	// for the current schema. Kept as a no-op for interface compliance.
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InstanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/compute/servers/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return // already gone, idempotent
		}
		resp.Diagnostics.AddError("delete instance", err.Error())
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
