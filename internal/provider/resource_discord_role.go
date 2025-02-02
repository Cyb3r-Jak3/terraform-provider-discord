package provider

import (
	"context"
	"fmt"
	"github.com/Cyb3r-Jak3/discord-terraform/internal/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordRoleResource{}
var _ resource.ResourceWithImportState = &DiscordRoleResource{}

func NewDiscordRoleResource() resource.Resource {
	return &DiscordRoleResource{}
}

type DiscordRoleResource struct {
	client *Context
}

func (r *DiscordRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *DiscordRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord Role Resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The role ID",
				Computed:            true,
			},
			"server_id": schema.StringAttribute{
				MarkdownDescription: "The server ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The role name",
				Required:            true,
			},
			"position": schema.Int64Attribute{
				MarkdownDescription: "The position of the role",
				Optional:            true,
				Default:             int64default.StaticInt64(1),
				Computed:            true,
			},
			"color": schema.Int64Attribute{
				MarkdownDescription: "The color of the role",
				Optional:            true,
			},
			"hoist": schema.BoolAttribute{
				MarkdownDescription: "Whether the role is hoisted",
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Computed:            true,
			},
			"mentionable": schema.BoolAttribute{
				MarkdownDescription: "Whether the role is mentionable",
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Computed:            true,
			},
			"managed": schema.BoolAttribute{
				MarkdownDescription: "Whether the role is managed",
				Computed:            true,
			},
			"permissions": schema.Int64Attribute{
				MarkdownDescription: "The permissions of the role",
				Optional:            true,
			},
		},
	}
}

func (r *DiscordRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DiscordRoleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	roleColor := int(data.Color.ValueInt64())
	role, err := client.GuildRoleCreate(data.ServerID.ValueString(), &discordgo.RoleParams{
		Name:        data.Name.ValueString(),
		Permissions: data.Permissions.ValueInt64Pointer(),
		Color:       &roleColor,
		Hoist:       data.Hoist.ValueBoolPointer(),
		Mentionable: data.Mentionable.ValueBoolPointer(),
	}, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create role", err.Error())
		return
	}
	data = DiscordRoleModel{
		ID:          types.StringValue(role.ID),
		ServerID:    data.ServerID,
		Name:        types.StringValue(role.Name),
		Position:    types.Int64Value(int64(role.Position)),
		Color:       types.Int64Value(int64(role.Color)),
		Permissions: types.Int64Value(role.Permissions),
		Hoist:       types.BoolValue(role.Hoist),
		Mentionable: types.BoolValue(role.Mentionable),
		Managed:     types.BoolValue(role.Managed),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DiscordRoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	role, err := utils.GetRole(ctx, client, data.ServerID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to fetch role: %s", data.ID.ValueString()), err.Error())
		return
	}
	data = DiscordRoleModel{
		ID:          types.StringValue(role.ID),
		ServerID:    data.ServerID,
		Name:        types.StringValue(role.Name),
		Position:    types.Int64Value(int64(role.Position)),
		Color:       types.Int64Value(int64(role.Color)),
		Permissions: types.Int64Value(role.Permissions),
		Hoist:       types.BoolValue(role.Hoist),
		Mentionable: types.BoolValue(role.Mentionable),
		Managed:     types.BoolValue(role.Managed),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DiscordRoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	if plan.Position.ValueInt64() != state.Position.ValueInt64() {
		planPositionInt := int(plan.Position.ValueInt64())
		var oldRole *discordgo.Role
		server, err := client.Guild(state.ServerID.ValueString(), discordgo.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to fetch server: %s", state.ServerID.ValueString()), err.Error())
			return
		}
		for _, role := range server.Roles {
			if role.Position == planPositionInt {
				oldRole = role
				break
			}
		}
		param := []*discordgo.Role{{ID: state.ID.ValueString(), Position: planPositionInt}}
		if oldRole != nil {
			param = append(param, &discordgo.Role{ID: oldRole.ID, Position: planPositionInt})
		}
		if _, err := client.GuildRoleReorder(state.ServerID.ValueString(), param, discordgo.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError("Failed to re-order roles", err.Error())
		} else {
			state.Position = plan.Position
		}
	}
	roleColor := int(state.Color.ValueInt64())
	role, err := client.GuildRoleEdit(state.ServerID.ValueString(), state.ID.ValueString(), &discordgo.RoleParams{
		Name:        state.Name.ValueString(),
		Permissions: state.Permissions.ValueInt64Pointer(),
		Color:       &roleColor,
		Hoist:       state.Hoist.ValueBoolPointer(),
		Mentionable: state.Mentionable.ValueBoolPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update role: %s", state.ID.ValueString()), err.Error())
		return
	}
	state = DiscordRoleModel{
		ID:          types.StringValue(role.ID),
		ServerID:    state.ServerID,
		Name:        types.StringValue(role.Name),
		Position:    types.Int64Value(int64(role.Position)),
		Color:       types.Int64Value(int64(role.Color)),
		Permissions: types.Int64Value(role.Permissions),
		Hoist:       types.BoolValue(role.Hoist),
		Mentionable: types.BoolValue(role.Mentionable),
		Managed:     types.BoolValue(role.Managed),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DiscordRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DiscordRoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	err := client.GuildRoleDelete(data.ServerID.ValueString(), data.ID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete role: %s", data.ID.ValueString()), err.Error())
		return
	}

}

func (r *DiscordRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idparts := strings.Split(req.ID, "/")
	if len(idparts) != 2 {
		resp.Diagnostics.AddError("error importing Discord Role", "invalid ID specified. Please specify the ID as \"server_id/role_id\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("server_id"), idparts[0],
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("id"), idparts[1],
	)...)
}
