// sws_plan — look up a flavor / plan by name (e.g. "m1.small").
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PlanDataSource struct{ client *Client }

type PlanDataModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	VCPUs types.Int64  `tfsdk:"vcpus"`
	RAM   types.Int64  `tfsdk:"ram_mb"`
	Disk  types.Int64  `tfsdk:"disk_gb"`
}

func NewPlanDataSource() datasource.DataSource { return &PlanDataSource{} }

func (d *PlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plan"
}

func (d *PlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("internal", fmt.Sprintf("expected *Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *PlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a compute plan (flavor) by name.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true},
			"name":    schema.StringAttribute{Required: true},
			"vcpus":   schema.Int64Attribute{Computed: true},
			"ram_mb":  schema.Int64Attribute{Computed: true},
			"disk_gb": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *PlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg PlanDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var all []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		VCPUs int64  `json:"vcpus"`
		RAM   int64  `json:"ram"`
		Disk  int64  `json:"disk"`
	}
	if err := d.client.Do("GET", "/api/compute/plans", nil, &all); err != nil {
		resp.Diagnostics.AddError("list plans", err.Error())
		return
	}
	want := cfg.Name.ValueString()
	for _, p := range all {
		if p.Name == want {
			cfg.ID = types.StringValue(p.ID)
			cfg.VCPUs = types.Int64Value(p.VCPUs)
			cfg.RAM = types.Int64Value(p.RAM)
			cfg.Disk = types.Int64Value(p.Disk)
			resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
			return
		}
	}
	resp.Diagnostics.AddError("plan not found", fmt.Sprintf("no plan with name %q", want))
}
