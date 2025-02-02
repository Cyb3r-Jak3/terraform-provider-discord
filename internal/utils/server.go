package utils

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polds/imgbase64"
)

// DiscordServerModel represents a Discord server. Used by both the data source and the resource.
type DiscordServerModel struct {
	ServerID                    types.String `tfsdk:"server_id"`
	Name                        types.String `tfsdk:"name"`
	Region                      types.String `tfsdk:"region"`
	DefaultMessageNotifications types.Int64  `tfsdk:"default_message_notifications"`
	VerificationLevel           types.Int64  `tfsdk:"verification_level"`
	ExplicitContentFilter       types.Int64  `tfsdk:"explicit_content_filter"`
	AfkTimeout                  types.Int64  `tfsdk:"afk_timeout"`
	IconURL                     types.String `tfsdk:"icon_url"`
	IconDataURI                 types.String `tfsdk:"icon_data_uri"`
	IconHash                    types.String `tfsdk:"icon_hash"`
	SplashUrl                   types.String `tfsdk:"splash_url"`
	SplashDataURI               types.String `tfsdk:"splash_data_uri"`
	SplashHash                  types.String `tfsdk:"splash_hash"`
	AfkChannelID                types.String `tfsdk:"afk_channel_id"`
	OwnerID                     types.String `tfsdk:"owner_id"`
}

func BuildServerResourceSchema(managed bool) map[string]schema.Attribute {
	base := map[string]schema.Attribute{

		"region": schema.StringAttribute{
			Description: "Region of the server.",
			Optional:    true,
			Computed:    true,
		},
		"default_message_notifications": schema.Int64Attribute{
			Description: "Default message notifications level.",
			Optional:    true,
			Computed:    true,
		},
		"verification_level": schema.Int64Attribute{
			Description: "Verification level.",
			Optional:    true,
			Computed:    true,
		},
		"explicit_content_filter": schema.Int64Attribute{
			Description: "Explicit content filter level.",
			Optional:    true,
			Computed:    true,
		},
		"afk_timeout": schema.Int64Attribute{
			Description: "AFK timeout in seconds.",
			Optional:    true,
			Computed:    true,
		},
		"icon_hash": schema.StringAttribute{
			Description: "Icon hash.",
			Optional:    true,
			Computed:    true,
		},
		"splash_hash": schema.StringAttribute{
			Description: "Splash hash.",
			Optional:    true,
			Computed:    true,
		},
		"afk_channel_id": schema.StringAttribute{
			Description: "AFK channel ID.",
			Optional:    true,
			Computed:    true,
		},
		"owner_id": schema.StringAttribute{
			Description: "Owner ID.",
			Optional:    true,
			Computed:    true,
		},
		"icon_url": schema.StringAttribute{
			Description: "Icon URL.",
			Optional:    true,
			Computed:    true,
		},
		"splash_url": schema.StringAttribute{
			Description: "Splash URL.",
			Optional:    true,
			Computed:    true,
		},
		"icon_data_uri": schema.StringAttribute{
			Description: "Icon data URI.",
			Optional:    true,
			Computed:    true,
		},
		"splash_data_uri": schema.StringAttribute{
			Description: "Splash data URI.",
			Optional:    true,
			Computed:    true,
		},
	}
	if managed {
		base["server_id"] = schema.StringAttribute{
			Description: "ID of the server. Only one of `server_id` or `name` can be set.",
			Required:    true,
		}
		base["name"] = schema.StringAttribute{
			Description: "Name of the server. Only one of `server_id` or `name` can be set.",
			Optional:    true,
		}
	} else {
		base["server_id"] = schema.StringAttribute{
			Description: "ID of the server.",
			Computed:    true,
		}
		base["name"] = schema.StringAttribute{
			Description: "Name of the server.",
			Required:    true,
		}
	}
	return base
}

// DiscordServerCreate creates a new Discord server. Used by both the server and managed server resource.
func DiscordServerCreate(client *discordgo.Session, ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DiscordServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server, err := client.GuildCreate(data.Name.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create a server", err.Error())
		return
	}

	guildParams := BuildGuildParams(data)
	server, err = client.GuildEdit(server.ID, guildParams, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update server", err.Error())
		return
	}
	for _, channel := range server.Channels {
		if _, err := client.ChannelDelete(channel.ID); err != nil {
			resp.Diagnostics.AddError("Failed to delete channel", err.Error())
			return
		}
	}
	// Update owner's ID if the specified one is not as same as default,
	// because we will receive "User is already owner" error if update to the same one.
	ownerID := server.OwnerID
	if data.OwnerID.ValueString() != "" {
		if data.OwnerID.ValueString() != ownerID {
			ownerID = data.OwnerID.ValueString()
		}
		server, err = client.GuildEdit(server.ID, &discordgo.GuildParams{OwnerID: ownerID}, discordgo.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update server owner", err.Error())
			return

		}
	}

	data = BuildServerModel(server)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// DiscordServerRead reads a Discord server. Used by both the server and managed server resource.
func DiscordServerRead(client *discordgo.Session, ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DiscordServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	server, err := client.Guild(data.ServerID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to get server %s", data.Name.ValueString()), err.Error())
		return

	}
	data = BuildServerModel(server)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// DiscordServerUpdate updates a Discord server. Used by both the server and managed server resource.
func DiscordServerUpdate(client *discordgo.Session, ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DiscordServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	guildParams := BuildGuildParams(data)

	server, err := client.GuildEdit(data.ServerID.ValueString(), guildParams, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update server", err.Error())
		return

	}
	data = BuildServerModel(server)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// DiscordServerDelete deletes a Discord server. Used by both the server and managed server resource.
func DiscordServerDelete(client *discordgo.Session, ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DiscordServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	if err := client.GuildDelete(data.ServerID.ValueString(), discordgo.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete server %s", data.Name.ValueString()), err.Error())
		return

	}

}

func BuildGuildParams(data *DiscordServerModel) *discordgo.GuildParams {
	icon := ""
	if data.IconURL.ValueString() != "" {
		icon = imgbase64.FromRemote(data.IconURL.ValueString())
	}
	if data.IconDataURI.ValueString() != "" {
		icon = data.IconDataURI.ValueString()
	}
	splash := ""
	if data.SplashUrl.ValueString() != "" {
		icon = imgbase64.FromRemote(data.SplashUrl.ValueString())
	}
	if data.SplashDataURI.ValueString() != "" {
		icon = data.SplashDataURI.ValueString()
	}
	verificationLevel := discordgo.VerificationLevel(data.VerificationLevel.ValueInt64())
	return &discordgo.GuildParams{
		Name:                        data.Name.ValueString(),
		Region:                      data.Region.ValueString(),
		VerificationLevel:           &verificationLevel,
		DefaultMessageNotifications: int(data.DefaultMessageNotifications.ValueInt64()),
		AfkChannelID:                data.AfkChannelID.ValueString(),
		AfkTimeout:                  int(data.AfkTimeout.ValueInt64()),
		Icon:                        icon,
		Splash:                      splash,
		OwnerID:                     data.OwnerID.ValueString(),
	}

}

func BuildServerModel(server *discordgo.Guild) *DiscordServerModel {
	return &DiscordServerModel{
		ServerID:                    types.StringValue(server.ID),
		Name:                        types.StringValue(server.Name),
		Region:                      types.StringValue(server.Region),
		DefaultMessageNotifications: types.Int64Value(int64(server.DefaultMessageNotifications)),
		VerificationLevel:           types.Int64Value(int64(server.VerificationLevel)),
		ExplicitContentFilter:       types.Int64Value(int64(server.ExplicitContentFilter)),
		AfkTimeout:                  types.Int64Value(int64(server.AfkTimeout)),
		AfkChannelID:                types.StringValue(server.AfkChannelID),
		IconHash:                    types.StringValue(server.Icon),
		SplashUrl:                   types.StringValue(server.Splash),
		SplashDataURI:               types.StringValue(server.Splash),
		SplashHash:                  types.StringValue(server.Splash),
		OwnerID:                     types.StringValue(server.OwnerID),
	}
}
