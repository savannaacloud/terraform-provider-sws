// sws_load_balancer — Octavia load balancer (LBaaS).
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

type LoadBalancerResource struct{ client *Client }
type LoadBalancerModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	VipSubnetID types.String `tfsdk:"vip_subnet_id"`
	Description types.String `tfsdk:"description"`
	VipAddress  types.String `tfsdk:"vip_address"`
	Status      types.String `tfsdk:"status"`
}

func NewLoadBalancerResource() resource.Resource { return &LoadBalancerResource{} }
func (r *LoadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}
func (r *LoadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *LoadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keep := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "An Octavia load balancer. Pair with sws_lb_listener, sws_lb_pool, sws_lb_member, sws_lb_health_monitor.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"name":          schema.StringAttribute{Required: true},
			"vip_subnet_id": schema.StringAttribute{Required: true, PlanModifiers: replace},
			"description":   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: keep},
			"vip_address":   schema.StringAttribute{Computed: true, PlanModifiers: keep},
			"status":        schema.StringAttribute{Computed: true, PlanModifiers: keep},
		},
	}
}
func (r *LoadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{
		"name":          plan.Name.ValueString(),
		"vip_subnet_id": plan.VipSubnetID.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	var got struct {
		ID                 string `json:"id"`
		VipAddress         string `json:"vip_address"`
		ProvisioningStatus string `json:"provisioning_status"`
		Description        string `json:"description"`
	}
	if err := r.client.Do("POST", "/api/orchestration/load-balancers", body, &got); err != nil {
		resp.Diagnostics.AddError("create load balancer", err.Error())
		return
	}
	// Poll for ACTIVE (Octavia builds amphora — can take 1-3 min).
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		var lb struct {
			VipAddress         string `json:"vip_address"`
			ProvisioningStatus string `json:"provisioning_status"`
		}
		if err := r.client.Do("GET", "/api/orchestration/load-balancers/"+got.ID, nil, &lb); err == nil {
			if lb.ProvisioningStatus == "ACTIVE" || lb.ProvisioningStatus == "ERROR" {
				got.VipAddress = lb.VipAddress
				got.ProvisioningStatus = lb.ProvisioningStatus
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
	plan.ID = types.StringValue(got.ID)
	plan.VipAddress = types.StringValue(got.VipAddress)
	plan.Status = types.StringValue(got.ProvisioningStatus)
	plan.Description = types.StringValue(got.Description)
	if got.ProvisioningStatus == "ERROR" {
		resp.Diagnostics.AddError("load balancer create failed", "entered ERROR state")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LoadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadBalancerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var got struct {
		ID                 string `json:"id"`
		Name               string `json:"name"`
		VipAddress         string `json:"vip_address"`
		VipSubnetID        string `json:"vip_subnet_id"`
		ProvisioningStatus string `json:"provisioning_status"`
		Description        string `json:"description"`
	}
	err := r.client.Do("GET", "/api/orchestration/load-balancers/"+state.ID.ValueString(), nil, &got)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("read load balancer", err.Error())
		return
	}
	state.Name = types.StringValue(got.Name)
	state.VipAddress = types.StringValue(got.VipAddress)
	state.Status = types.StringValue(got.ProvisioningStatus)
	state.Description = types.StringValue(got.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
func (r *LoadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LoadBalancerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := map[string]any{"name": plan.Name.ValueString()}
	if !plan.Description.IsNull() {
		body["description"] = plan.Description.ValueString()
	}
	_ = r.client.Do("PUT", "/api/orchestration/load-balancers/"+plan.ID.ValueString(), body, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *LoadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadBalancerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// cascade=true tells Octavia to drop listeners/pools/members with the LB.
	err := r.client.Do("DELETE", "/api/orchestration/load-balancers/"+state.ID.ValueString()+"?cascade=true", nil, nil)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.Status == 404 {
			return
		}
		resp.Diagnostics.AddError("delete load balancer", err.Error())
	}
}
func (r *LoadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
