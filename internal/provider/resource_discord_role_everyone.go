package provider

import (
	"context"
	"fmt"
	"github.com/Cyb3r-Jak3/discord-terraform/internal/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordEveryoneRoleResource{}
var _ resource.ResourceWithImportState = &DiscordEveryoneRoleResource{}

func NewDiscordEveryoneRoleResource() resource.Resource {
	return &DiscordEveryoneRoleResource{}
}

type DiscordEveryoneRoleResource struct {
	client *Context
}

type DiscordEveryoneRoleModel struct {
	ServerID    types.String `tfsdk:"server_id"`
	Permissions types.Int64  `tfsdk:"permissions"`
}

func (r *DiscordEveryoneRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_everyone"
}

func (r *DiscordEveryoneRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"permissions": schema.Int64Attribute{
				MarkdownDescription: "The permissions of the role",
				Optional:            true,
				Default:             int64default.StaticInt64(0),
				Computed:            true,
			},
		},
	}
}

func (r *DiscordEveryoneRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordEveryoneRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DiscordEveryoneRoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	serverID := data.ServerID.ValueString()
	role, err := utils.GetRole(ctx, client, serverID, serverID)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to everyone role for server %s", serverID), err.Error())
		return
	}
	data = &DiscordEveryoneRoleModel{
		ServerID:    data.ServerID,
		Permissions: types.Int64Value(role.Permissions),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordEveryoneRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DiscordEveryoneRoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	serverID := data.ServerID.ValueString()
	role, err := utils.GetRole(ctx, client, serverID, serverID)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to everyone role for server %s", serverID), err.Error())
		return
	}
	data = &DiscordEveryoneRoleModel{
		ServerID:    data.ServerID,
		Permissions: types.Int64Value(role.Permissions),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordEveryoneRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DiscordRoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	serverID := state.ServerID.ValueString()
	newPermissions := plan.Permissions.ValueInt64()
	role, err := client.GuildRoleEdit(serverID, serverID, &discordgo.RoleParams{
		Permissions: &newPermissions,
	})
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update everyone role for server %s", state.ID.ValueString()), err.Error())
		return
	}
	state.Permissions = types.Int64Value(role.Permissions)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DiscordEveryoneRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DiscordRoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("Cannot delete @everyone role", "Cannot delete @everyone role. If you need to remove the resource then delete it from your state")

}

func (r *DiscordEveryoneRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idparts := strings.Split(req.ID, "/")
	if len(idparts) != 2 {
		resp.Diagnostics.AddError("error importing Discord Message", "invalid ID specified. Please specify the ID as \"server_id/role_id\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("server_id"), idparts[0],
	)...)
}
