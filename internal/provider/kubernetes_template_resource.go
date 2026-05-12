// sws_kubernetes_template — Magnum cluster template (reusable blueprint).
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

type KubernetesTemplateResource struct{ client *Client }
type KubernetesTemplateModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Image             types.String `tfsdk:"image"`
	KeypairID         types.String `tfsdk:"keypair_id"`
	ExternalNetworkID types.String `tfsdk:"external_network_id"`
	DNSNameserver     types.String `tfsdk:"dns_nameserver"`
	MasterFlavorID    types.String `tfsdk:"master_flavor_id"`
	FlavorID          types.String `tfsdk:"flavor_id"`
	CoeName           types.String `tfsdk:"coe_name"`
}

func NewKubernetesTemplateResource() resource.Resource { return &KubernetesTemplateResource{} }
func (r *KubernetesTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_template"
}
func (r *KubernetesTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *KubernetesTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Magnum cluster template. All attributes are RequiresReplace; templates are typically read-only after creation.",
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":                schema.StringAttribute{Required: true, PlanModifiers: replace},
			"image":               schema.StringAttribute{Required: true, Description: "fcos image UUID/name", PlanModifiers: replace},
			"keypair_id":          schema.StringAttribute{Required: true, PlanModifiers: replace},
			"external_network_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"dns_nameserver":      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: append(keep, replace...)},
			"master_flavor_id":    schema.StringAttribute{Required: true, PlanModifiers: replace},
			"flavor_id":           schema.StringAttribute{Required: true, Description: "worker flavor", PlanModifiers: replace},
			"coe_name":            schema.StringAttribute{Optional: true, Computed: true, Description: "kubernetes (default), swarm, mesos", PlanModifiers: append(keep, replace...)},
		},
	}
}
func (r *KubernetesTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KubernetesTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	coe := "kubernetes"
	if !plan.CoeName.IsNull() && !plan.CoeName.IsUnknown() && plan.CoeName.ValueString() != "" {
		coe = plan.CoeName.ValueString()
	}
	dns := "8.8.8.8"
	if !plan.DNSNameserver.IsNull() && !plan.DNSNameserver.IsUnknown() && plan.DNSNameserver.ValueString() != "" {
		dns = plan.DNSNameserver.ValueString()
	}
	body := map[string]any{
		"name":                plan.Name.ValueString(),
		"image_id":            plan.Image.ValueString(),
		"keypair_id":          plan.KeypairID.ValueString(),
		"external_network_id": plan.ExternalNetworkID.ValueString(),
		"dns_nameserver":      dns,
		"master_flavor_id":    plan.MasterFlavorID.ValueString(),
		"flavor_id":           plan.FlavorID.ValueString(),
		"coe":                 coe,
	}
	var got struct {
		ID string `json:"id"`
	}
	if err := r.client.Do("POST", "/api/orchestration/cluster-templates", body, &got); err != nil {
		resp.Diagnostics.AddError("create k8s template", err.Error())
		return
	}
	plan.ID = types.StringValue(got.ID)
	plan.CoeName = types.StringValue(coe)
	plan.DNSNameserver = types.StringValue(dns)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *KubernetesTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KubernetesTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID string `json:"id"`
	}
	err := r.client.Do("GET", "/api/orchestration/cluster-templates/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *KubernetesTemplateResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}
func (r *KubernetesTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KubernetesTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/cluster-templates/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete k8s template", err.Error())
	}
}
func (r *KubernetesTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
