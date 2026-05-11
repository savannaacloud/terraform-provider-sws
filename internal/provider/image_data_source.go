// sws_image — look up a Glance image by name. Resolves to a UUID
// users can plug into sws_instance.image without hard-coding it.
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ImageDataSource struct{ client *Client }

type ImageDataModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewImageDataSource() datasource.DataSource { return &ImageDataSource{} }

func (d *ImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (d *ImageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a Savannaa OS image by name.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"name": schema.StringAttribute{Required: true},
		},
	}
}

func (d *ImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg ImageDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var all []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := d.client.Do("GET", "/api/compute/images", nil, &all); err != nil {
		resp.Diagnostics.AddError("list images", err.Error())
		return
	}
	want := cfg.Name.ValueString()
	for _, img := range all {
		if img.Name == want {
			cfg.ID = types.StringValue(img.ID)
			resp.Diagnostics.Append(resp.State.Set(ctx, cfg)...)
			return
		}
	}
	resp.Diagnostics.AddError("image not found", fmt.Sprintf("no image with name %q (try one of: %v)", want, imageNames(all)))
}

func imageNames(imgs []struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}) []string {
	out := make([]string, 0, len(imgs))
	for _, i := range imgs {
		out = append(out, i.Name)
	}
	return out
}
