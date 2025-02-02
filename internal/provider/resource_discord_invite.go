package provider

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordInvite{}
var _ resource.ResourceWithImportState = &DiscordInvite{}

func NewDiscordInviteResource() resource.Resource {
	return &DiscordInvite{}
}

type DiscordInvite struct {
	client *Context
}

type DiscordInviteModel struct {
	ChannelID types.String `tfsdk:"channel_id"`
	MaxAge    types.Int64  `tfsdk:"max_age"`
	MaxUses   types.Int64  `tfsdk:"max_uses"`
	Temporary types.Bool   `tfsdk:"temporary"`
	Unique    types.Bool   `tfsdk:"unique"`
	Code      types.String `tfsdk:"code"`
	ID        types.String `tfsdk:"id"`
}

func (r *DiscordInvite) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (r *DiscordInvite) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord Invite Resource",

		Attributes: map[string]schema.Attribute{
			"channel_id": schema.StringAttribute{
				Description: "The channel ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_age": schema.Int64Attribute{
				Description: "The duration in seconds that the invite will be valid for",
				Optional:    true,
				Default:     int64default.StaticInt64(86400),
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"max_uses": schema.Int64Attribute{
				Description: "The maximum number of times the invite can be used",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"temporary": schema.BoolAttribute{
				Description: "Whether the invite grants temporary membership",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Computed: true,
			},
			"unique": schema.BoolAttribute{
				Description: "Whether the invite is unique",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Computed: true,
			},
			"code": schema.StringAttribute{
				Description: "The invite code",
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description:        "The invite code",
				Computed:           true,
				DeprecationMessage: "Use the `code` attribute instead",
			},
		},
	}
}

func (r *DiscordInvite) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordInvite) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DiscordInviteModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session

	channelId := data.ChannelID.ValueString()
	invite, err := client.ChannelInviteCreate(channelId, discordgo.Invite{
		MaxAge:    int(data.MaxAge.ValueInt64()),
		MaxUses:   int(data.MaxUses.ValueInt64()),
		Temporary: data.Temporary.ValueBool(),
		Unique:    data.Unique.ValueBool(),
	}, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create a invite", err.Error())
		return
	}
	data = &DiscordInviteModel{
		ChannelID: types.StringValue(channelId),
		MaxAge:    types.Int64Value(int64(invite.MaxAge)),
		MaxUses:   types.Int64Value(int64(invite.MaxUses)),
		Temporary: types.BoolValue(invite.Temporary),
		Unique:    types.BoolValue(invite.Unique),
		Code:      types.StringValue(invite.Code),
		ID:        types.StringValue(invite.Code),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordInvite) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DiscordInviteModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	// In order to get the invite metadata, we need to get all the invites for the channel
	// and then filter the invite we want by the code
	invites, err := client.ChannelInvites(data.ChannelID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get invites from API", err.Error())
		return
	}
	var invite *discordgo.Invite
	for _, i := range invites {
		if i.Code == data.Code.ValueString() {
			invite = i
			break
		}
	}
	if invite == nil {
		resp.Diagnostics.AddError("Failed to get invite", "Invite not found in the channel")
		return
	}
	data = &DiscordInviteModel{
		ChannelID: types.StringValue(data.ChannelID.ValueString()),
		MaxAge:    types.Int64Value(int64(invite.MaxAge)),
		MaxUses:   types.Int64Value(int64(invite.MaxUses)),
		Temporary: types.BoolValue(invite.Temporary),
		Unique:    types.BoolValue(invite.Unique),
		Code:      types.StringValue(invite.Code),
		ID:        types.StringValue(invite.Code),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordInvite) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DiscordInviteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("failed to update discord invite", "Not implemented")
}

func (r *DiscordInvite) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DiscordInviteModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	if _, err := client.InviteDelete(data.Code.ValueString(), discordgo.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError("Failed to delete invite", err.Error())
		return
	}
}

func (r *DiscordInvite) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idparts := strings.Split(req.ID, "/")
	if len(idparts) != 2 {
		resp.Diagnostics.AddError("error importing Discord Invite", "invalid ID specified. Please specify the ID as \"channel_id/invite_code\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("channel_id"), idparts[0],
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("code"), idparts[1],
	)...)
}
