package provider

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DiscordRoleDatasource{}

func NewDiscordRoleDataSource() datasource.DataSource {
	return &DiscordRoleDatasource{}
}

type DiscordRoleModel struct {
	ServerID    types.String `tfsdk:"server_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Position    types.Int64  `tfsdk:"position"`
	Color       types.Int64  `tfsdk:"color"`
	Permissions types.Int64  `tfsdk:"permissions"`
	Hoist       types.Bool   `tfsdk:"hoist"`
	Mentionable types.Bool   `tfsdk:"mentionable"`
	Managed     types.Bool   `tfsdk:"managed"`
}

type DiscordRoleDatasource struct {
	client *Context
}

func (r *DiscordRoleDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"

}

func (r *DiscordRoleDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *DiscordRoleDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Discord Role",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
			},
			"position": schema.Int64Attribute{
				Computed: true,
			},
			"color": schema.Int64Attribute{
				Computed: true,
			},
			"permissions": schema.Int64Attribute{
				Computed: true,
			},
			"hoist": schema.BoolAttribute{
				Computed: true,
			},
			"mentionable": schema.BoolAttribute{
				Computed: true,
			},
			"managed": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (r *DiscordRoleDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *DiscordRoleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	roles, err := client.GuildRoles(data.ServerID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to fetch role %s", data.ID.ValueString()), err.Error())
		return
	}
	var selectedRole *discordgo.Role
	for _, role := range roles {
		if role.ID == data.ID.ValueString() || role.Name == data.Name.ValueString() {
			selectedRole = role
			break
		}
	}
	if selectedRole == nil {
		resp.Diagnostics.AddError("Role not found", "Role not found in the server")
		return
	}
	data = &DiscordRoleModel{
		ServerID:    types.StringValue(data.ServerID.ValueString()),
		ID:          types.StringValue(selectedRole.ID),
		Name:        types.StringValue(selectedRole.Name),
		Position:    types.Int64Value(int64(selectedRole.Position)),
		Color:       types.Int64Value(int64(selectedRole.Color)),
		Permissions: types.Int64Value(selectedRole.Permissions),
		Hoist:       types.BoolValue(selectedRole.Hoist),
		Mentionable: types.BoolValue(selectedRole.Mentionable),
		Managed:     types.BoolValue(selectedRole.Managed),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
