package provider

import (
	"context"
	"fmt"
	"github.com/Cyb3r-Jak3/discord-terraform/internal/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &DiscordServerDatasource{}

func NewDiscordServerDataSource() datasource.DataSource {
	return &DiscordServerDatasource{}
}

type DiscordServerDatasource struct {
	client *Context
}

func (r *DiscordServerDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"

}

func (r *DiscordServerDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *DiscordServerDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord Server Data Source.\n This data source can only fetch up to 1000 servers.",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Description: "ID of the server. Only one of `server_id` or `name` can be set.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the server. Only one of `server_id` or `name` can be set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("server_id")),
				},
			},
			"region": schema.StringAttribute{
				Description: "Region of the server.",
				Computed:    true,
			},
			"default_message_notifications": schema.Int64Attribute{
				Description: "Default message notifications level.",
				Computed:    true,
			},
			"verification_level": schema.Int64Attribute{
				Description: "Verification level.",
				Computed:    true,
			},
			"explicit_content_filter": schema.Int64Attribute{
				Description: "Explicit content filter level.",
				Computed:    true,
			},
			"afk_timeout": schema.Int64Attribute{
				Description: "AFK timeout in seconds.",
				Computed:    true,
			},
			"icon_hash": schema.StringAttribute{
				Description: "Icon hash.",
				Computed:    true,
			},
			"splash_hash": schema.StringAttribute{
				Description: "Splash hash.",
				Computed:    true,
			},
			"afk_channel_id": schema.StringAttribute{
				Description: "AFK channel ID.",
				Computed:    true,
			},
			"owner_id": schema.StringAttribute{
				Description: "Owner ID.",
				Computed:    true,
			},
			"icon_url": schema.StringAttribute{
				Description: "Icon URL.",
				Computed:    true,
			},
			"splash_url": schema.StringAttribute{
				Description: "Splash URL.",
				Computed:    true,
			},
			"icon_data_uri": schema.StringAttribute{
				Description: "Icon Data URI.",
				Computed:    true,
			},
			"splash_data_uri": schema.StringAttribute{
				Description: "Splash Data URI.",
				Computed:    true,
			},
		},
	}
}

func (r *DiscordServerDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *utils.DiscordServerModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var guild *discordgo.Guild
	var err error

	client := r.client.Session
	serverID := data.ServerID.ValueString()
	serverName := data.Name.ValueString()
	if serverID != "" {
		guild, err = client.Guild(serverID, discordgo.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("failed to get server: %s", serverID), err.Error())
			return
		}
	} else {
		guilds, guildListErr := client.UserGuilds(1000, "", "", false, discordgo.WithContext(ctx))
		if guildListErr != nil {
			resp.Diagnostics.AddError("failed to fetch servers", guildListErr.Error())
			return
		}
		for _, g := range guilds {
			if g.Name == serverName {
				guild, err = client.Guild(g.ID, discordgo.WithContext(ctx))
				if err != nil {
					resp.Diagnostics.AddError(fmt.Sprintf("failed to get server: %s", serverName), err.Error())
					return
				}
				break
			}
		}
		if guild == nil {
			resp.Diagnostics.AddError(fmt.Sprintf("failed to get server: %s", serverName), "server not found")
			return
		}
	}
	data = utils.BuildServerModel(guild)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
