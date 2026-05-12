// sws_kubernetes_cluster — Magnum cluster instance.
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

type KubernetesClusterResource struct{ client *Client }
type KubernetesClusterModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ClusterTemplateID types.String `tfsdk:"cluster_template_id"`
	NodeCount         types.Int64  `tfsdk:"node_count"`
	MasterCount       types.Int64  `tfsdk:"master_count"`
	KeypairID         types.String `tfsdk:"keypair_id"`
	Status            types.String `tfsdk:"status"`
	APIAddress        types.String `tfsdk:"api_address"`
}

func NewKubernetesClusterResource() resource.Resource { return &KubernetesClusterResource{} }
func (r *KubernetesClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}
func (r *KubernetesClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *KubernetesClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Magnum Kubernetes cluster. Polls for CREATE_COMPLETE (typically 8-15 min).",
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":                schema.StringAttribute{Required: true, PlanModifiers: replace},
			"cluster_template_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"node_count":          schema.Int64Attribute{Required: true},
			"master_count":        schema.Int64Attribute{Optional: true, Computed: true, Description: "default 1"},
			"keypair_id":          schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: append(keep, replace...)},
			"status":              schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"api_address":         schema.StringAttribute{Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *KubernetesClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KubernetesClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mc := int64(1)
	if !plan.MasterCount.IsNull() && !plan.MasterCount.IsUnknown() {
		mc = plan.MasterCount.ValueInt64()
	}
	body := map[string]any{
		"name":                plan.Name.ValueString(),
		"cluster_template_id": plan.ClusterTemplateID.ValueString(),
		"node_count":          plan.NodeCount.ValueInt64(),
		"master_count":        mc,
	}
	if !plan.KeypairID.IsNull() && !plan.KeypairID.IsUnknown() && plan.KeypairID.ValueString() != "" {
		body["keypair"] = plan.KeypairID.ValueString()
	}
	var got struct {
		UUID string `json:"uuid"`
		ID   string `json:"id"`
	}
	if err := r.client.Do("POST", "/api/orchestration/clusters", body, &got); err != nil {
		resp.Diagnostics.AddError("create k8s cluster", err.Error())
		return
	}
	id := got.UUID
	if id == "" {
		id = got.ID
	}
	// Poll for CREATE_COMPLETE (typically 8-15 min for Magnum).
	deadline := time.Now().Add(20 * time.Minute)
	var status, api string
	for time.Now().Before(deadline) {
		var c struct {
			Status     string `json:"status"`
			APIAddress string `json:"api_address"`
			MasterCount int64 `json:"master_count"`
		}
		if err := r.client.Do("GET", "/api/orchestration/clusters/"+id, nil, &c); err == nil {
			status = c.Status
			api = c.APIAddress
			if c.MasterCount > 0 {
				mc = c.MasterCount
			}
			if status == "CREATE_COMPLETE" || status == "CREATE_FAILED" {
				break
			}
		}
		time.Sleep(30 * time.Second)
	}
	plan.ID = types.StringValue(id)
	plan.Status = types.StringValue(status)
	plan.APIAddress = types.StringValue(api)
	plan.MasterCount = types.Int64Value(mc)
	if status == "CREATE_FAILED" {
		resp.Diagnostics.AddError("k8s cluster create failed", "cluster entered CREATE_FAILED state")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *KubernetesClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		Status      string `json:"status"`
		APIAddress  string `json:"api_address"`
		NodeCount   int64  `json:"node_count"`
		MasterCount int64  `json:"master_count"`
	}
	err := r.client.Do("GET", "/api/orchestration/clusters/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read k8s cluster", err.Error())
		return
	}
	state.Status = types.StringValue(got.Status)
	state.APIAddress = types.StringValue(got.APIAddress)
	state.NodeCount = types.Int64Value(got.NodeCount)
	state.MasterCount = types.Int64Value(got.MasterCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *KubernetesClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan KubernetesClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Magnum only really supports resize (node_count change).
	body := map[string]any{"node_count": plan.NodeCount.ValueInt64()}
	_ = r.client.Do("PUT", "/api/orchestration/clusters/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *KubernetesClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KubernetesClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do("DELETE", "/api/orchestration/clusters/"+state.ID.ValueString(), nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete k8s cluster", err.Error())
	}
}
func (r *KubernetesClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
