// Package provider implements the Savannaa Cloud Terraform provider.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SwsProvider satisfies the provider.Provider interface.
type SwsProvider struct {
	version string
}

// SwsProviderModel mirrors the HCL block:
//
//	provider "sws" {
//	  api_url      = "https://api.savannaa.com/v3"
//	  api_key      = "ctk_..."
//	  project_name = "user-..."
//	  region       = "ng-abuja-1"
//	}
//
// Every field is optional in HCL; missing fields fall back to env vars
// (SWS_API_URL, SWS_API_KEY, SWS_PROJECT_NAME, SWS_REGION). This matches
// how kubectl, gcloud, and the openstack CLI all handle credentials so
// users don't have to commit secrets into .tf files.
type SwsProviderModel struct {
	APIURL      types.String `tfsdk:"api_url"`
	APIKey      types.String `tfsdk:"api_key"`
	ProjectName types.String `tfsdk:"project_name"`
	Region      types.String `tfsdk:"region"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SwsProvider{version: version}
	}
}

func (p *SwsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sws"
	resp.Version = p.version
}

func (p *SwsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for Savannaa Cloud. Manages compute instances, networks, keypairs, and security groups via the Savannaa API.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Savannaa API base URL. Defaults to env var `SWS_API_URL`, then `https://api.savannaa.com/v3`.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key generated from **Account → API Keys** in the console (starts with `ctk_`). Defaults to env var `SWS_API_KEY`. Marked sensitive so it doesn't leak into Terraform output.",
				Optional:            true,
				Sensitive:           true,
			},
			"project_name": schema.StringAttribute{
				MarkdownDescription: "Project name to operate against. Each account has a default `user-<username>` project. Defaults to env var `SWS_PROJECT_NAME`.",
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region: `ng-abuja-1` or `ng-lagos-1`. Defaults to env var `SWS_REGION`, then `ng-abuja-1`.",
				Optional:            true,
			},
		},
	}
}

func (p *SwsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SwsProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := firstSet(data.APIURL.ValueString(), os.Getenv("SWS_API_URL"), "https://api.savannaa.com/v3")
	apiKey := firstSet(data.APIKey.ValueString(), os.Getenv("SWS_API_KEY"))
	project := firstSet(data.ProjectName.ValueString(), os.Getenv("SWS_PROJECT_NAME"))
	region := firstSet(data.Region.ValueString(), os.Getenv("SWS_REGION"), "ng-abuja-1")

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API key",
			"Set the `api_key` provider attribute or the SWS_API_KEY env var. Generate one in the console: Account → API Keys.",
		)
		return
	}

	client := NewClient(apiURL, apiKey, project, region)
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *SwsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
		NewKeypairResource,
		NewNetworkResource,
		NewSecurityGroupResource,
	}
}

func (p *SwsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewImageDataSource,
		NewPlanDataSource,
	}
}

func firstSet(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
