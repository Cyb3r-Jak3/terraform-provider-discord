package provider

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordSystemChannelResource{}
var _ resource.ResourceWithImportState = &DiscordSystemChannelResource{}

func NewDiscordSystemChannelResource() resource.Resource {
	return &DiscordEveryoneRoleResource{}
}

type DiscordSystemChannelResource struct {
	client *Context
}

type DiscordSystemChannelResourceModel struct {
	ServerID        types.String `tfsdk:"server_id"`
	SystemChannelID types.String `tfsdk:"permissions"`
}

func (r *DiscordSystemChannelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_channel"
}

func (r *DiscordSystemChannelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord @everyone Role Resource",

		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				MarkdownDescription: "The server ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"system_channel_id": schema.StringAttribute{
				MarkdownDescription: "Channel ID for system messages",
				Required:            true,
			},
		},
	}
}

func (r *DiscordSystemChannelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordSystemChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DiscordSystemChannelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	serverID := data.ServerID.ValueString()
	server, err := client.Guild(serverID, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to find server %s", serverID), err.Error())
		return
	}
	systemChannelID := server.SystemChannelID
	if systemChannelID != data.SystemChannelID.ValueString() {
		systemChannelID = data.SystemChannelID.ValueString()
	}
	if _, err := client.GuildEdit(serverID, &discordgo.GuildParams{
		SystemChannelID: systemChannelID,
	}); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update system channel id server %s", serverID), err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordSystemChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DiscordSystemChannelResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	serverID := data.ServerID.ValueString()
	server, err := client.Guild(serverID, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error fetching server %s", serverID), err.Error())
		return

	}
	data.SystemChannelID = types.StringValue(server.SystemChannelID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordSystemChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DiscordSystemChannelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	client := r.client.Session
	server, err := client.Guild(data.ServerID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error fetching server %s", data.ServerID.ValueString()), err.Error())
		return

	}
	if data.SystemChannelID.ValueString() != server.SystemChannelID {
		if _, err := client.GuildEdit(data.ServerID.ValueString(), &discordgo.GuildParams{
			SystemChannelID: data.SystemChannelID.ValueString(),
		}); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to update system channel id server %s", data.ServerID.ValueString()), err.Error())
			return
		}

	}

}

func (r *DiscordSystemChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DiscordSystemChannelResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("Cannot delete @everyone role", "Cannot delete @everyone role. If you need to remove the resource then delete it from your state")

}

func (r *DiscordSystemChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idparts := strings.Split(req.ID, "/")
	if len(idparts) != 2 {
		resp.Diagnostics.AddError("error importing Discord Message", "invalid ID specified. Please specify the ID as \"server_id/role_id\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("server_id"), idparts[0],
	)...)
}
