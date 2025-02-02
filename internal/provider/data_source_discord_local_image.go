package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polds/imgbase64"
)

var _ datasource.DataSource = &DiscordLocalFile{}

func NewDiscordLocalImageDataSource() datasource.DataSource {
	return &DiscordLocalFile{}
}

type DiscordLocalImageModel struct {
	File    types.String `tfsdk:"file"`
	DataURI types.String `tfsdk:"data_uri"`
}

type DiscordLocalFile struct {
	client *Context
}

func (r *DiscordLocalFile) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_local_image"

}

func (r *DiscordLocalFile) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Context)

	if !ok {
		resp.Diagnostics.AddError(
			"unexpected resource configure type",
			fmt.Sprintf("Expected *Context, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *DiscordLocalFile) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Discord Color Data Source",
		Attributes: map[string]schema.Attribute{
			"file": schema.StringAttribute{
				Description: "The file to read",
				Required:    true,
			},
			"data_uri": schema.StringAttribute{
				Description: "The data uri of the file",
				Computed:    true,
			},
		},
	}
}

func (r *DiscordLocalFile) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DiscordLocalImageModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	file := data.File.ValueString()
	if img, err := imgbase64.FromLocal(file); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to process %s", file), err.Error())
		return
	} else {
		data.DataURI = types.StringValue(img)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
