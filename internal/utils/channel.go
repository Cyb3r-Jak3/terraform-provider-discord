package utils

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DiscordChannelModel struct {
	ServerID              types.String `tfsdk:"server_id"`
	ChannelID             types.String `tfsdk:"channel_id"`
	Category              types.String `tfsdk:"category"`
	Type                  types.String `tfsdk:"type"`
	Name                  types.String `tfsdk:"name"`
	Position              types.Int64  `tfsdk:"position"`
	SyncPermsWithCategory types.Bool   `tfsdk:"sync_perms_with_category,omitempty"`
	Topic                 types.String `tfsdk:"topic"`
	NSFW                  types.Bool   `tfsdk:"nsfw"`
	Bitrate               types.Int64  `tfsdk:"bitrate,omitempty"`
	UserLimit             types.Int64  `tfsdk:"user_limit,omitempty"`
}

func DiscordChannelCreate(client *discordgo.Session, ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DiscordChannelModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	channelParams, err := BuildChannelParams(data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build channel params", err.Error())
		return

	}
	channel, err := client.GuildChannelCreateComplex(data.ServerID.ValueString(), channelParams, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create a channel", err.Error())
		return
	}
	data, err = BuildChannelModel(channel)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build channel model", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func DiscordChannelRead(client *discordgo.Session, ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DiscordChannelModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := client.Channel(data.ChannelID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch channel", err.Error())
		return
	}

	data, err = BuildChannelModel(channel)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build channel model", err.Error())
		return
	}
	if data.Type.ValueString() != "category" {
		if channel.ParentID == "" {
			data.SyncPermsWithCategory = types.BoolValue(false)
		} else {
			parent, err := client.Channel(channel.ParentID)
			if err != nil {
				resp.Diagnostics.AddError("Failed to fetch category of channel", err.Error())
				return
			}

			data.SyncPermsWithCategory = types.BoolValue(ArePermissionsSynced(channel, parent))
		}

	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func DiscordChannelUpdate(client *discordgo.Session, ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DiscordChannelModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := client.Channel(data.ChannelID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch channel", err.Error())
		return
	}

	channelParams, err := BuildChannelParams(data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build channel params", err.Error())
		return

	}
	if data.SyncPermsWithCategory.ValueBool() {
		if channel.ParentID == "" {
			resp.Diagnostics.AddError("Channel does not have a category", "")
			return
		}

		parent, err := client.Channel(channel.ParentID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to fetch category of channel", err.Error())
			return
		}

		if !ArePermissionsSynced(channel, parent) {
			if err := SyncChannelPermissions(client, ctx, parent, channel); err != nil {
				resp.Diagnostics.AddError("Failed to sync permissions with category", err.Error())
				return
			}
		}
	}

	channel, err = client.ChannelEditComplex(data.ChannelID.ValueString(), &discordgo.ChannelEdit{
		Name:      channelParams.Name,
		Position:  &channelParams.Position,
		Topic:     channelParams.Topic,
		NSFW:      &channelParams.NSFW,
		Bitrate:   channelParams.Bitrate,
		UserLimit: channelParams.UserLimit,
		ParentID:  channelParams.ParentID,
	}, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update channel", err.Error())
		return
	}

	data, err = BuildChannelModel(channel)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build channel model", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func DiscordChannelDelete(client *discordgo.Session, ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DiscordChannelModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := client.ChannelDelete(data.ChannelID.ValueString(), discordgo.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError("Failed to delete channel", err.Error())
		return
	}

}

func BuildChannelParams(data *DiscordChannelModel) (discordgo.GuildChannelCreateData, error) {
	channelType, okay := GetDiscordChannelType(data.Type.ValueString())
	if !okay {
		return discordgo.GuildChannelCreateData{}, fmt.Errorf("invalid channel type: %s", data.Type.ValueString())
	}
	return discordgo.GuildChannelCreateData{
		Name:      data.Name.ValueString(),
		Position:  int(data.Position.ValueInt64()),
		Type:      channelType,
		Topic:     data.Topic.ValueString(),
		NSFW:      data.NSFW.ValueBool(),
		Bitrate:   int(data.Bitrate.ValueInt64()),
		UserLimit: int(data.UserLimit.ValueInt64()),
		ParentID:  data.Category.ValueString(),
	}, nil

}

func BuildChannelModel(channel *discordgo.Channel) (*DiscordChannelModel, error) {
	channelType, okay := GetTextChannelType(channel.Type)
	if !okay {
		return &DiscordChannelModel{}, fmt.Errorf("invalid channel type: %s", channelType)
	}

	return &DiscordChannelModel{
		ServerID:  types.StringValue(channel.GuildID),
		ChannelID: types.StringValue(channel.ID),
		Category:  types.StringValue(channel.ParentID),
		Type:      types.StringValue(channelType),
		Name:      types.StringValue(channel.Name),
		Position:  types.Int64Value(int64(channel.Position)),
		Topic:     types.StringValue(channel.Topic),
		NSFW:      types.BoolValue(channel.NSFW),
		Bitrate:   types.Int64Value(int64(channel.Bitrate)),
		UserLimit: types.Int64Value(int64(channel.UserLimit)),
	}, nil
}

func BuildChannelResourceSchema(channelType string) map[string]schema.Attribute {
	base := map[string]schema.Attribute{
		"server_id": schema.StringAttribute{
			Description: "The server ID",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			Description: "The channel name",
			Required:    true,
		},
		"channel_id": schema.StringAttribute{
			Description: "The channel ID",
			Computed:    true,
		},
		"type": schema.StringAttribute{
			Description: "The channel type",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.OneOf(channelType),
			},
			DeprecationMessage: "This field is deprecated. Type is now inferred from the resource name.",
		},
		"position": schema.Int64Attribute{
			Description: "Sorting position of the channel",
			Optional:    true,
			Default:     int64default.StaticInt64(1),
			Computed:    true,
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
		},
	}
	if channelType != "category" {
		base["sync_perms_with_category"] = schema.BoolAttribute{
			Description: "Whether to sync permissions with the category",
			Optional:    true,
		}
		base["category"] = schema.StringAttribute{
			Description: "The category ID",
			Optional:    true,
		}
	}
	if Contains([]string{"text", "news"}, channelType) {
		base["topic"] = schema.StringAttribute{
			Description: "The channel topic",
			Optional:    true,
		}

	}
	if channelType == "voice" {
		base["bitrate"] = schema.Int64Attribute{
			Description: "The bitrate of the channel",
			Optional:    true,
			Default:     int64default.StaticInt64(64000),
			Computed:    true,
		}
		base["user_limit"] = schema.Int64Attribute{
			Description: "The user limit of the channel",
			Optional:    true,
		}

	}
	if channelType == "text" {
		base["nsfw"] = schema.BoolAttribute{
			Description: "Whether the channel is NSFW",
			Optional:    true,
			Default:     booldefault.StaticBool(false),
			Computed:    true,
		}

	}
	return base
}

func GetChannelTypeText(channelType discordgo.ChannelType) (string, bool) {
	switch channelType {
	case 0:
		return "text", true
	case 2:
		return "voice", true
	case 4:
		return "category", true
	case 5:
		return "news", true
	case 6:
		return "store", true
	}

	return "text", false
}

func GetDiscordChannelType(name string) (discordgo.ChannelType, bool) {
	switch name {
	case "text":
		return discordgo.ChannelTypeGuildText, true
	case "voice":
		return discordgo.ChannelTypeGuildVoice, true
	case "category":
		return discordgo.ChannelTypeGuildCategory, true
	case "news":
		return discordgo.ChannelTypeGuildNews, true
	case "store":
		return discordgo.ChannelTypeGuildStore, true
	}

	return 0, false
}

func ArePermissionsSynced(from *discordgo.Channel, to *discordgo.Channel) bool {
	for _, p1 := range from.PermissionOverwrites {
		cont := false
		for _, p2 := range to.PermissionOverwrites {
			if p1.ID == p2.ID && p1.Type == p2.Type && p1.Allow == p2.Allow && p1.Deny == p2.Deny {
				cont = true
				break
			}
		}
		if !cont {
			return false
		}
	}

	for _, p1 := range to.PermissionOverwrites {
		cont := false
		for _, p2 := range from.PermissionOverwrites {
			if p1.ID == p2.ID && p1.Type == p2.Type && p1.Allow == p2.Allow && p1.Deny == p2.Deny {
				cont = true
				break
			}
		}
		if !cont {
			return false
		}
	}

	return true
}

func SyncChannelPermissions(c *discordgo.Session, ctx context.Context, from *discordgo.Channel, to *discordgo.Channel) error {
	for _, p := range to.PermissionOverwrites {
		if err := c.ChannelPermissionDelete(to.ID, p.ID); err != nil {
			return err
		}
	}

	for _, p := range from.PermissionOverwrites {
		if err := c.ChannelPermissionSet(to.ID, p.ID, discordgo.PermissionOverwriteTypeRole, p.Allow, p.Deny, discordgo.WithContext(ctx)); err != nil {
			return err
		}
	}

	return nil
}
