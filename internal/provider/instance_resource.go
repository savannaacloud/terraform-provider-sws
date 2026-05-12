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


// isUUID returns true if s looks like an OpenStack/Glance UUID
// (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
			if !isHex {
				return false
			}
		}
	}
	return true
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

	// Backend expects flavor_id + image_id (UUIDs) per the launch payload
	// the web console sends. The HCL accepts either a UUID or a plan name;
	// resolve the name → UUID on the fly so users can keep using the
	// idiomatic data.sws_plan.NAME.name reference.
	flavorID := plan.Plan.ValueString()
	if !isUUID(flavorID) {
		var all []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := r.client.Do("GET", "/api/compute/plans", nil, &all); err != nil {
			resp.Diagnostics.AddError("resolve plan", err.Error())
			return
		}
		found := ""
		for _, pl := range all {
			if pl.Name == flavorID {
				found = pl.ID
				break
			}
		}
		if found == "" {
			resp.Diagnostics.AddError("resolve plan", fmt.Sprintf("no plan with name %q", flavorID))
			return
		}
		flavorID = found
	}

	body := map[string]any{
		"name":      plan.Name.ValueString(),
		"flavor_id": flavorID,
		"image_id":  plan.Image.ValueString(),
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
	var status string
	for time.Now().Before(deadline) {
		var got serverGetResponse
		if err := r.client.Do("GET", "/api/compute/servers/"+created.ID, nil, &got); err == nil {
			if got.Status == "ACTIVE" || got.Status == "ERROR" {
				status = got.Status
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
	if status == "" {
		resp.Diagnostics.AddError("instance create timed out", "server did not become ACTIVE within 5 min")
		return
	}
	if status == "ERROR" {
		plan.ID = types.StringValue(created.ID)
		plan.Status = types.StringValue("ERROR")
		plan.IPAddress = types.StringValue("")
		resp.Diagnostics.AddError("instance create failed", "server entered ERROR state")
		return
	}

	// Auto-allocate a public IP if the user asked for one. The backend does
	// this as a separate post-ACTIVE call (matches the launch wizard flow).
	// Failure here isn't fatal — log a warning so the user can attach one
	// manually, but keep the instance.
	if !plan.PublicIP.IsNull() && plan.PublicIP.ValueBool() {
		if err := r.client.Do("POST", "/api/compute/servers/"+created.ID+"/auto-public-ip", nil, nil); err != nil {
			resp.Diagnostics.AddWarning("auto public IP failed", err.Error()+" — instance is up; you can attach a public IP manually.")
		}
	}

	// Re-fetch to get the addresses dict populated (including the floating
	// IP if auto-public-ip just allocated one). Poll briefly because the
	// allocation can take a couple of seconds to land in the server view.
	var final serverGetResponse
	for i := 0; i < 6; i++ {
		if err := r.client.Do("GET", "/api/compute/servers/"+created.ID, nil, &final); err == nil {
			pubIP, fixedIP := extractIPs(final.Addresses)
			if pubIP != "" || !plan.PublicIP.ValueBool() {
				// done: got the floating IP, OR caller didn't ask for one
				plan.ID = types.StringValue(final.ID)
				plan.Status = types.StringValue(final.Status)
				ip := pubIP
				if ip == "" {
					ip = fixedIP
				}
				plan.IPAddress = types.StringValue(ip)
				resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
				return
			}
		}
		time.Sleep(3 * time.Second)
	}
	// Fall back to whatever we have — better than no state.
	pubIP, fixedIP := extractIPs(final.Addresses)
	ip := pubIP
	if ip == "" {
		ip = fixedIP
	}
	plan.ID = types.StringValue(created.ID)
	plan.Status = types.StringValue(final.Status)
	plan.IPAddress = types.StringValue(ip)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// serverGetResponse mirrors the python-backend _fmt_server() shape.
// addresses is a {network_name: [{addr, OS-EXT-IPS:type, ...}, ...]} dict.
type serverGetResponse struct {
	ID        string                     `json:"id"`
	Name      string                     `json:"name"`
	Status    string                     `json:"status"`
	Addresses map[string][]map[string]any `json:"addresses"`
}

// extractIPs returns (floatingIP, fixedIP) from the backend addresses dict.
// Prefers the first floating IPv4 across all networks; falls back to first
// fixed IPv4 if no floating is present.
func extractIPs(addrs map[string][]map[string]any) (string, string) {
	var pub, fixed string
	for _, ifaces := range addrs {
		for _, ip := range ifaces {
			ver, _ := ip["version"].(float64)
			if ver != 4 && ver != 0 {
				continue
			}
			addr, _ := ip["addr"].(string)
			if addr == "" {
				continue
			}
			t, _ := ip["OS-EXT-IPS:type"].(string)
			if t == "floating" && pub == "" {
				pub = addr
			}
			if t != "floating" && fixed == "" {
				fixed = addr
			}
		}
	}
	return pub, fixed
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InstanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got serverGetResponse
	if state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
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
	pub, fixed := extractIPs(got.Addresses)
	ip := pub
	if ip == "" {
		ip = fixed
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
