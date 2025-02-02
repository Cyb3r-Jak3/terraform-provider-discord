package provider

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DiscordMember{}

func NewDiscordMemberDataSource() datasource.DataSource {
	return &DiscordMember{}
}

type DiscordMemberModel struct {
	ServerID      types.String `tfsdk:"server_id"`
	UserID        types.String `tfsdk:"user_id"`
	Username      types.String `tfsdk:"username"`
	Discriminator types.String `tfsdk:"discriminator"`
	JoinedAt      types.String `tfsdk:"joined_at"`
	PremiumSince  types.String `tfsdk:"premium_since"`
	Avatar        types.String `tfsdk:"avatar"`
	Nick          types.String `tfsdk:"nick"`
	Roles         types.Set    `tfsdk:"roles"`
	InServer      types.Bool   `tfsdk:"in_server"`
}

type DiscordMember struct {
	client *Context
}

func (r *DiscordMember) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_member"

}

func (r *DiscordMember) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *DiscordMember) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Discord Member Data Source",
		Attributes: map[string]schema.Attribute{
			"server_id": schema.StringAttribute{
				Description: "The ID of the server to search for the member in.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to search for. Only one of `user_id` or `username` can be set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("username")),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username of the user to search for. Only one of `user_id` or `username` can be set.",
				Optional:    true,
			},
			"discriminator": schema.StringAttribute{
				Description:        "The discriminator of the user to search for. Required if `username` is set.",
				Optional:           true,
				DeprecationMessage: "Discriminator is being deprecated by Discord. Only use this if there are users who haven't migrated their username.",
			},
			"joined_at": schema.StringAttribute{
				Description: "The date and time the user joined the server.",
				Computed:    true,
			},
			"premium_since": schema.StringAttribute{
				Description: "The date and time the user started boosting the server.",
				Computed:    true,
			},
			"avatar": schema.StringAttribute{
				Description: "The URL of the user's avatar.",
				Computed:    true,
			},
			"nick": schema.StringAttribute{
				Description: "The user's nickname in the server.",
				Computed:    true,
			},
			"roles": schema.SetAttribute{
				Description: "The roles the user has in the server.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"in_server": schema.BoolAttribute{
				Description: "Whether the user is in the server.",
				Computed:    true,
			},
		},
	}
}

func (r *DiscordMember) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DiscordMemberModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var member *discordgo.Member
	var memberErr error

	client := r.client.Session
	serverId := data.ServerID.ValueString()
	userID := data.UserID.ValueString()

	if userID != "" {
		member, memberErr = client.GuildMember(serverId, userID, discordgo.WithContext(ctx))
		if memberErr != nil {
			resp.Diagnostics.AddError("failed to get member", memberErr.Error())
			return
		}
	} else {
		username := data.Username.ValueString()
		members, err := client.GuildMembersSearch(serverId, username, 1, discordgo.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("failed to fetch members for %s", serverId), err.Error())
			return
		}
		discriminator := data.Discriminator.ValueString()
		for _, m := range members {
			if m.User.Username == username && m.User.Discriminator == discriminator {
				member = m
				break
			}
		}
	}

	if member == nil {
		resp.Diagnostics.AddError("member not found", "no member found with the given criteria")
		return
	}
	roles := make([]attr.Value, 0, len(member.Roles))
	for _, role := range member.Roles {
		roles = append(roles, types.StringValue(role))
	}

	data = DiscordMemberModel{
		ServerID:      types.StringValue(serverId),
		UserID:        types.StringValue(member.User.ID),
		Username:      types.StringValue(member.User.Username),
		Discriminator: types.StringValue(member.User.Discriminator),
		JoinedAt:      types.StringValue(member.JoinedAt.String()),
		Nick:          types.StringValue(member.Nick),
		Roles:         types.SetValueMust(types.StringType, roles),
	}
	if member.PremiumSince != nil {
		data.PremiumSince = types.StringValue(member.PremiumSince.String())
	}
	if member.User != nil {
		data.Avatar = types.StringValue(member.User.AvatarURL(""))
		data.InServer = types.BoolValue(true)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
