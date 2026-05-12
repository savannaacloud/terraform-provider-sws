// sws_keypair — Nova SSH keypair for injecting public keys into instances.
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

type KeypairResource struct{ client *Client }

type KeypairModel struct {
	Name       types.String `tfsdk:"name"`
	PublicKey  types.String `tfsdk:"public_key"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

func NewKeypairResource() resource.Resource { return &KeypairResource{} }

func (r *KeypairResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

func (r *KeypairResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeypairResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	immutable := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "SSH keypair Savannaa injects into new instances at boot. The `id` is the keypair name — Nova uses name, not UUID, as the primary key.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Keypair name (also the resource ID).",
				PlanModifiers:       immutable,
			},
			"public_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "OpenSSH public key (e.g. `ssh-rsa AAAA...`).",
				PlanModifiers:       immutable,
			},
			"fingerprint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "MD5 fingerprint Nova returned.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *KeypairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KeypairModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"name":       plan.Name.ValueString(),
		"public_key": plan.PublicKey.ValueString(),
	}
	var got struct {
		Name        string `json:"name"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := r.client.Do("POST", "/api/compute/keypairs", body, &got); err != nil {
		resp.Diagnostics.AddError("create keypair", err.Error())
		return
	}
	plan.Fingerprint = types.StringValue(got.Fingerprint)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KeypairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KeypairModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Backend has no GET /keypairs/{name} (only POST + LIST + DELETE).
	// LIST + filter by suffix; the backend prefixes stored keypair names
	// with a per-project tag, so the stored name ends with the user-given
	// name. Match suffix-with-dash, plus bare-name for legacy rows.
	var all []struct {
		Name        string `json:"name"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := r.client.Do("GET", "/api/compute/keypairs", nil, &all); err != nil {
		resp.Diagnostics.AddError("read keypair", err.Error())
		return
	}
	want := state.Name.ValueString()
	var got *struct {
		Name        string `json:"name"`
		Fingerprint string `json:"fingerprint"`
	}
	for i := range all {
		n := all[i].Name
		if n == want || (len(n) > len(want) && n[len(n)-len(want)-1:] == "-"+want) {
			got = &all[i]
			break
		}
	}
	if got == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Fingerprint = types.StringValue(got.Fingerprint)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *KeypairResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All fields are RequiresReplace; this is unreachable.
}

func (r *KeypairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KeypairModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/compute/keypairs/"+state.Name.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete keypair", err.Error())
	}
}

func (r *KeypairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
