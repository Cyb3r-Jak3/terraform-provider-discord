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
			},
			"unique": schema.BoolAttribute{
				Description: "Whether the invite is unique",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"code": schema.StringAttribute{
				Description: "The invite code",
				Computed:    true,
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
	invite, err := client.Invite(data.Code.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to get invite: %s", data.Code.ValueString()), err.Error())
		return
	}
	data = &DiscordInviteModel{
		ChannelID: types.StringValue(invite.Channel.ID),
		MaxAge:    types.Int64Value(int64(invite.MaxAge)),
		MaxUses:   types.Int64Value(int64(invite.MaxUses)),
		Temporary: types.BoolValue(invite.Temporary),
		Unique:    types.BoolValue(invite.Unique),
		Code:      types.StringValue(invite.Code),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordInvite) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DiscordInviteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("failed to discord invite", "Not implemented")
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
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
}
