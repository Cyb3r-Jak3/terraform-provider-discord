package provider

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polds/imgbase64"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordWebhook{}
var _ resource.ResourceWithImportState = &DiscordWebhook{}

func NewDiscordWebhookResource() resource.Resource {
	return &DiscordWebhook{}
}

type DiscordWebhook struct {
	client *Context
}

type DiscordWebhookModel struct {
	ID            types.String `tfsdk:"id"`
	ChannelID     types.String `tfsdk:"channel_id"`
	GuildID       types.String `tfsdk:"guild_id"`
	Name          types.String `tfsdk:"name"`
	AvatarURL     types.String `tfsdk:"avatar_url"`
	AvatarDataURI types.String `tfsdk:"avatar_data_uri"`
	AvatarHash    types.String `tfsdk:"avatar_hash"`
	Token         types.String `tfsdk:"token"`
	URL           types.String `tfsdk:"url"`
	SlackURL      types.String `tfsdk:"slack_url"`
	GithubURL     types.String `tfsdk:"github_url"`
}

func (r *DiscordWebhook) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *DiscordWebhook) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordWebhook) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord Webhook Resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The webhook ID",
				Computed:            true,
			},
			"channel_id": schema.StringAttribute{
				Description: "The channel ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"guild_id": schema.StringAttribute{
				Description: "The guild ID",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The webhook name",
				Required:    true,
			},
			"avatar_url": schema.StringAttribute{
				Description: "The URL of the avatar.\nIf this attribute is set then you will not be able to import the resource.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("avatar_url"),
						path.MatchRoot("avatar_data_uri"),
					),
				},
			},
			"avatar_data_uri": schema.StringAttribute{
				Description: "The data URI of the avatar.\nIf this attribute is set then you will not be able to import the resource.",
				Optional:    true,
				Computed:    true,
			},
			"avatar_hash": schema.StringAttribute{
				Description: "The hash of the avatar",
				Computed:    true,
			},
			"token": schema.StringAttribute{
				Description: "The token of the webhook",
				Computed:    true,
				Sensitive:   true,
			},
			"url": schema.StringAttribute{
				Description: "The URL of the webhook",
				Computed:    true,
				Sensitive:   true,
			},
			"slack_url": schema.StringAttribute{
				Description: "The Slack URL of the webhook",
				Computed:    true,
				Sensitive:   true,
			},
			"github_url": schema.StringAttribute{
				Description: "The GitHub URL of the webhook",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *DiscordWebhook) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DiscordWebhookModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session

	channelId := data.ChannelID.ValueString()
	avatar := ""
	if data.AvatarURL.ValueString() != "" {
		avatar = imgbase64.FromRemote(data.AvatarURL.ValueString())
	} else if data.AvatarDataURI.ValueString() != "" {
		avatar = data.AvatarDataURI.ValueString()
	}
	webhook, err := client.WebhookCreate(channelId, data.Name.ValueString(), avatar, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create a webhook", err.Error())
		return
	}
	webhookURL := fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", webhook.ID, webhook.Token)
	data = &DiscordWebhookModel{
		ID:            types.StringValue(webhook.ID),
		ChannelID:     types.StringValue(channelId),
		Name:          types.StringValue(data.Name.ValueString()),
		GuildID:       types.StringValue(webhook.GuildID),
		AvatarURL:     types.StringValue(data.AvatarURL.ValueString()),
		AvatarDataURI: types.StringValue(data.AvatarDataURI.ValueString()),
		AvatarHash:    types.StringValue(webhook.Avatar),
		Token:         types.StringValue(webhook.Token),
		URL:           types.StringValue(webhookURL),
		SlackURL:      types.StringValue(webhookURL + "/slack"),
		GithubURL:     types.StringValue(webhookURL + "/github"),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordWebhook) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DiscordWebhookModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	webhook, err := client.Webhook(data.ID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to get webhook %s", data.ID.ValueString()), err.Error())
		return
	}
	webhookURL := fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", webhook.ID, webhook.Token)
	data = &DiscordWebhookModel{
		ID:            types.StringValue(webhook.ID),
		ChannelID:     types.StringValue(webhook.ChannelID),
		Name:          types.StringValue(webhook.Name),
		GuildID:       types.StringValue(webhook.GuildID),
		AvatarURL:     types.StringValue(data.AvatarURL.ValueString()),
		AvatarDataURI: types.StringValue(data.AvatarDataURI.ValueString()),
		AvatarHash:    types.StringValue(webhook.Avatar),
		Token:         types.StringValue(webhook.Token),
		URL:           types.StringValue(webhookURL),
		SlackURL:      types.StringValue(webhookURL + "/slack"),
		GithubURL:     types.StringValue(webhookURL + "/github"),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordWebhook) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DiscordWebhookModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session

	channelId := data.ChannelID.ValueString()
	avatar := ""
	if data.AvatarURL.ValueString() != "" {
		avatar = imgbase64.FromRemote(data.AvatarURL.ValueString())
	} else if data.AvatarDataURI.ValueString() != "" {
		avatar = data.AvatarDataURI.ValueString()
	}
	webhook, err := client.WebhookEdit(data.ID.ValueString(), data.Name.ValueString(), avatar, channelId, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update webhook %s", data.ID.ValueString()), err.Error())
		return
	}

	webhookURL := fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", webhook.ID, webhook.Token)
	data = &DiscordWebhookModel{
		ID:            types.StringValue(webhook.ID),
		ChannelID:     types.StringValue(webhook.ChannelID),
		GuildID:       types.StringValue(webhook.GuildID),
		Name:          types.StringValue(data.Name.ValueString()),
		AvatarURL:     types.StringValue(data.AvatarURL.ValueString()),
		AvatarDataURI: types.StringValue(data.AvatarDataURI.ValueString()),
		AvatarHash:    types.StringValue(webhook.Avatar),
		Token:         types.StringValue(webhook.Token),
		URL:           types.StringValue(webhookURL),
		SlackURL:      types.StringValue(webhookURL + "/slack"),
		GithubURL:     types.StringValue(webhookURL + "/github"),
	}

	resp.Diagnostics.AddError("failed to discord invite", "Not implemented")
}

func (r *DiscordWebhook) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DiscordWebhookModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	if err := client.WebhookDelete(data.ID.ValueString(), discordgo.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete webhook %s", data.ID.ValueString()), err.Error())
		return
	}
}

func (r *DiscordWebhook) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
