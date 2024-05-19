package provider

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DiscordSystemChannel{}

func NewDiscordSystemChannelDataSource() datasource.DataSource {
	return &DiscordSystemChannel{}
}

type DiscordSystemChannelModel struct {
	ServerID        types.String `tfsdk:"server_id"`
	SystemChannelID types.String `tfsdk:"system_channel_id"`
}

type DiscordSystemChannel struct {
	client *Context
}

func (r *DiscordSystemChannel) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_channel"

}

func (r *DiscordSystemChannel) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *DiscordSystemChannel) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Discord System Channel Data Source",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Description: "The server ID",
				Required:    true,
			},
			"system_channel_id": schema.StringAttribute{
				Description: "The system channel ID",
				Computed:    true,
			},
		},
	}
}

func (r *DiscordSystemChannel) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DiscordSystemChannelModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	var err error
	var server *discordgo.Guild

	client := r.client.Session
	serverID := data.ServerID.ValueString()
	server, err = client.Guild(serverID)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to get server: %s", serverID), err.Error())
		return
	}
	data = DiscordSystemChannelModel{
		ServerID:        types.StringValue(serverID),
		SystemChannelID: types.StringValue(server.SystemChannelID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
