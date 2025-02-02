package provider

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordMessage{}
var _ resource.ResourceWithImportState = &DiscordMessage{}

func NewDiscordMessageResource() resource.Resource {
	return &DiscordMessage{}
}

type DiscordMessage struct {
	client *Context
}

type DiscordMessageModel struct {
	ID              types.String               `tfsdk:"id"`
	ChannelID       types.String               `tfsdk:"channel_id"`
	ServerID        types.String               `tfsdk:"server_id"`
	AuthorID        types.String               `tfsdk:"author_id"`
	Content         types.String               `tfsdk:"content"`
	Timestamp       types.String               `tfsdk:"timestamp"`
	EditedTimestamp types.String               `tfsdk:"edited_timestamp"`
	TTS             types.Bool                 `tfsdk:"tts"`
	Pinned          types.Bool                 `tfsdk:"pinned"`
	Type            types.Int64                `tfsdk:"type"`
	Embed           []DiscordMessageEmbedModel `tfsdk:"embed"`
}

type DiscordMessageEmbedModel struct {
	Title       types.String                       `tfsdk:"title"`
	Description types.String                       `tfsdk:"description"`
	URL         types.String                       `tfsdk:"url"`
	Timestamp   types.String                       `tfsdk:"timestamp"`
	Color       types.Int64                        `tfsdk:"color"`
	Footer      *DiscordMessageEmbedFooterModel    `tfsdk:"footer"`
	Image       *DiscordMessageEmbedImageModel     `tfsdk:"image"`
	Thumbnail   *DiscordMessageEmbedThumbnailModel `tfsdk:"thumbnail"`
	Video       *DiscordMessageEmbedVideoModel     `tfsdk:"video"`
	Provider    *DiscordMessageEmbedProviderModel  `tfsdk:"provider"`
	Author      *DiscordMessageEmbedAuthorModel    `tfsdk:"author"`
	Fields      []*DiscordMessageEmbedFieldModel   `tfsdk:"fields"`
}

type DiscordMessageEmbedFooterModel struct {
	Text         types.String `tfsdk:"text"`
	IconURL      types.String `tfsdk:"icon_url"`
	ProxyIconURL types.String `tfsdk:"proxy_icon_url"`
}

type DiscordMessageEmbedImageModel struct {
	URL      types.String `tfsdk:"url"`
	ProxyURL types.String `tfsdk:"proxy_url"`
	Height   types.Int64  `tfsdk:"height"`
	Width    types.Int64  `tfsdk:"width"`
}

type DiscordMessageEmbedThumbnailModel struct {
	URL      types.String `tfsdk:"url"`
	ProxyURL types.String `tfsdk:"proxy_url"`
	Height   types.Int64  `tfsdk:"height"`
	Width    types.Int64  `tfsdk:"width"`
}

type DiscordMessageEmbedVideoModel struct {
	URL    types.String `tfsdk:"url"`
	Height types.Int64  `tfsdk:"height"`
	Width  types.Int64  `tfsdk:"width"`
}

type DiscordMessageEmbedProviderModel struct {
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
}

type DiscordMessageEmbedAuthorModel struct {
	Name         types.String `tfsdk:"name"`
	URL          types.String `tfsdk:"url"`
	IconURL      types.String `tfsdk:"icon_url"`
	ProxyIconURL types.String `tfsdk:"proxy_icon_url"`
}

type DiscordMessageEmbedFieldModel struct {
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
	Inline types.Bool   `tfsdk:"inline"`
}

func (r *DiscordMessage) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_message"
}

func (r *DiscordMessage) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord Message Resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The message ID",
				Computed:            true,
			},
			"channel_id": schema.StringAttribute{
				MarkdownDescription: "The channel ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				MarkdownDescription: "The server ID",
				Computed:            true,
			},
			"author_id": schema.StringAttribute{
				MarkdownDescription: "The author ID",
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The message content",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 2000),
					stringvalidator.ExactlyOneOf(path.MatchRoot("content"), path.MatchRoot("embed")),
				},
			},
			"timestamp": schema.StringAttribute{
				MarkdownDescription: "The timestamp of the message",
				Computed:            true,
			},
			"edited_timestamp": schema.StringAttribute{
				MarkdownDescription: "The timestamp of the last edit",
				Optional:            true,
				Computed:            true,
			},
			"tts": schema.BoolAttribute{
				MarkdownDescription: "Whether the message is TTS",
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Computed:            true,
			},
			"pinned": schema.BoolAttribute{
				MarkdownDescription: "Whether the message is pinned",
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				Computed:            true,
			},
			"type": schema.Int64Attribute{
				MarkdownDescription: "The message type",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"embed": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"title": schema.StringAttribute{
							MarkdownDescription: "The title of the embed",
							Optional:            true,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "The description of the embed",
							Optional:            true,
							Computed:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "The URL of the embed",
							Optional:            true,
							Computed:            true,
						},
						"timestamp": schema.StringAttribute{
							MarkdownDescription: "The timestamp of the embed",
							Optional:            true,
							Computed:            true,
						},
						"color": schema.Int64Attribute{
							MarkdownDescription: "The color of the embed",
							Optional:            true,
							Computed:            true,
						},
					},
					Blocks: map[string]schema.Block{
						"footer": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"text": schema.StringAttribute{
									MarkdownDescription: "The text of the footer",
									Optional:            true,
									Computed:            true,
								},
								"icon_url": schema.StringAttribute{
									MarkdownDescription: "The icon URL of the footer",
									Optional:            true,
									Computed:            true,
								},
								"proxy_icon_url": schema.StringAttribute{
									MarkdownDescription: "The proxy icon URL of the footer",
									Computed:            true,
								},
							},
						},
						"image": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"url": schema.StringAttribute{
									MarkdownDescription: "The URL of the image",
									Optional:            true,
									Computed:            true,
								},
								"proxy_url": schema.StringAttribute{
									MarkdownDescription: "The proxy URL of the image",
									Computed:            true,
								},
								"height": schema.Int64Attribute{
									MarkdownDescription: "The height of the image",
									Optional:            true,
									Computed:            true,
								},
								"width": schema.Int64Attribute{
									MarkdownDescription: "The width of the image",
									Optional:            true,
									Computed:            true,
								},
							},
						},
						"thumbnail": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"url": schema.StringAttribute{
									MarkdownDescription: "The URL of the thumbnail",
									Optional:            true,
									Computed:            true,
								},
								"proxy_url": schema.StringAttribute{
									MarkdownDescription: "The proxy URL of the thumbnail",
									Computed:            true,
								},
								"height": schema.Int64Attribute{
									MarkdownDescription: "The height of the thumbnail",
									Optional:            true,
									Computed:            true,
								},
								"width": schema.Int64Attribute{
									MarkdownDescription: "The width of the thumbnail",
									Optional:            true,
									Computed:            true,
								},
							},
						},
						"video": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"url": schema.StringAttribute{
									MarkdownDescription: "The URL of the video",
									Optional:            true,
									Computed:            true,
								},
								"height": schema.Int64Attribute{
									MarkdownDescription: "The height of the video",
									Optional:            true,
								},
								"width": schema.Int64Attribute{
									MarkdownDescription: "The width of the video",
									Optional:            true,
								},
							},
						},
						"provider": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the provider",
									Optional:            true,
								},
								"url": schema.StringAttribute{
									MarkdownDescription: "The URL of the provider",
									Optional:            true,
								},
							},
						},
						"author": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the author",
									Optional:            true,
									Computed:            true,
								},
								"url": schema.StringAttribute{
									MarkdownDescription: "The URL of the author",
									Optional:            true,
									Computed:            true,
								},
								"icon_url": schema.StringAttribute{
									MarkdownDescription: "The icon URL of the author",
									Optional:            true,
									Computed:            true,
								},
								"proxy_icon_url": schema.StringAttribute{
									MarkdownDescription: "The proxy icon URL of the author",
									Computed:            true,
								},
							},
						},
						"fields": schema.SetNestedBlock{
							Validators: []validator.Set{
								setvalidator.SizeAtMost(25),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "The name of the field",
										Required:            true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "The value of the field",
										Optional:            true,
									},
									"inline": schema.BoolAttribute{
										MarkdownDescription: "Whether the field is inline",
										Optional:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *DiscordMessage) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordMessage) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DiscordMessageModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	embeds := buildEmbedMessages(data)
	message, err := client.ChannelMessageSendComplex(data.ChannelID.ValueString(), &discordgo.MessageSend{
		Content: data.Content.ValueString(),
		Embeds:  embeds,
		TTS:     data.TTS.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to send message", err.Error())
		return
	}
	if data.Pinned.ValueBool() {
		err = client.ChannelMessagePin(data.ChannelID.ValueString(), message.ID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to pin message", err.Error())
			return
		}
	}
	data = DiscordMessageModel{
		ID:        types.StringValue(message.ID),
		ChannelID: types.StringValue(message.ChannelID),
		ServerID:  types.StringValue(message.GuildID),
		AuthorID:  types.StringValue(message.Author.ID),
		Content:   types.StringValue(message.Content),
		Timestamp: types.StringValue(message.Timestamp.Format("2006-01-02T15:04:05.000Z")),
		TTS:       types.BoolValue(message.TTS),
		Pinned:    types.BoolValue(data.Pinned.ValueBool()),
		Type:      types.Int64Value(int64(message.Type)),
		Embed:     unbuildEmbedMessages(message.Embeds),
	}
	if message.EditedTimestamp != nil {
		data.EditedTimestamp = types.StringValue(message.EditedTimestamp.Format("2006-01-02T15:04:05.000Z"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordMessage) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DiscordMessageModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	message, err := client.ChannelMessage(data.ChannelID.ValueString(), data.ID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to get message", err.Error())
		return
	}
	data = DiscordMessageModel{
		ID:        types.StringValue(message.ID),
		ChannelID: types.StringValue(message.ChannelID),
		ServerID:  types.StringValue(message.GuildID),
		AuthorID:  types.StringValue(message.Author.ID),
		Content:   types.StringValue(message.Content),
		Timestamp: types.StringValue(message.Timestamp.Format("2006-01-02T15:04:05.000Z")),
		TTS:       types.BoolValue(message.TTS),
		Pinned:    types.BoolValue(message.Pinned),
		Type:      types.Int64Value(int64(message.Type)),
		Embed:     unbuildEmbedMessages(message.Embeds),
	}
	if message.EditedTimestamp != nil {
		data.EditedTimestamp = types.StringValue(message.EditedTimestamp.Format("2006-01-02T15:04:05.000Z"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordMessage) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DiscordMessageModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	embeds := buildEmbedMessages(data)
	message, err := client.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: data.ChannelID.ValueString(),
		ID:      data.ID.ValueString(),
		Content: data.Content.ValueStringPointer(),
		Embeds:  &embeds,
	}, discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update message", err.Error())
		return
	}
	data.ID = types.StringValue(message.ID)
	if data.Pinned.ValueBool() {
		err = client.ChannelMessagePin(data.ChannelID.ValueString(), message.ID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to pin message", err.Error())
			return
		}
	}
	data = DiscordMessageModel{
		ID:              types.StringValue(message.ID),
		ChannelID:       types.StringValue(message.ChannelID),
		ServerID:        types.StringValue(message.GuildID),
		AuthorID:        types.StringValue(message.Author.ID),
		Content:         types.StringValue(message.Content),
		TTS:             types.BoolValue(message.TTS),
		Pinned:          types.BoolValue(data.Pinned.ValueBool()),
		Type:            types.Int64Value(int64(message.Type)),
		Embed:           unbuildEmbedMessages(message.Embeds),
		EditedTimestamp: types.StringValue(message.EditedTimestamp.Format("2006-01-02T15:04:05.000Z")),
	}
}

func (r *DiscordMessage) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DiscordMessageModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	err := client.ChannelMessageDelete(data.ChannelID.ValueString(), data.ID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete message", err.Error())
		return
	}

}

func (r *DiscordMessage) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idparts := strings.Split(req.ID, "/")
	if len(idparts) != 2 {
		resp.Diagnostics.AddError("error importing Discord Message", "invalid ID specified. Please specify the ID as \"channel_id/message_id\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("channel_id"), idparts[0],
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("id"), idparts[1],
	)...)
}

func buildEmbedMessages(data DiscordMessageModel) []*discordgo.MessageEmbed {
	embeds := make([]*discordgo.MessageEmbed, 0, len(data.Embed))
	for _, embed := range data.Embed {
		base := &discordgo.MessageEmbed{
			Title:       embed.Title.ValueString(),
			Description: embed.Description.ValueString(),
			URL:         embed.URL.ValueString(),
			Timestamp:   embed.Timestamp.ValueString(),
			Color:       int(embed.Color.ValueInt64()),
		}
		for _, field := range embed.Fields {
			base.Fields = append(base.Fields, &discordgo.MessageEmbedField{
				Name:   field.Name.ValueString(),
				Value:  field.Value.ValueString(),
				Inline: field.Inline.ValueBool(),
			})
		}
		if embed.Image != nil {
			base.Image = &discordgo.MessageEmbedImage{
				URL:      embed.Image.URL.ValueString(),
				ProxyURL: embed.Image.ProxyURL.ValueString(),
				Height:   int(embed.Image.Height.ValueInt64()),
				Width:    int(embed.Image.Width.ValueInt64()),
			}
		}
		if embed.Thumbnail != nil {
			base.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL:      embed.Thumbnail.URL.ValueString(),
				ProxyURL: embed.Thumbnail.ProxyURL.ValueString(),
				Height:   int(embed.Thumbnail.Height.ValueInt64()),
				Width:    int(embed.Thumbnail.Width.ValueInt64()),
			}
		}
		if embed.Video != nil {
			base.Video = &discordgo.MessageEmbedVideo{
				URL:    embed.Video.URL.ValueString(),
				Height: int(embed.Video.Height.ValueInt64()),
				Width:  int(embed.Video.Width.ValueInt64()),
			}
		}
		if embed.Footer != nil {
			base.Footer = &discordgo.MessageEmbedFooter{
				Text:         embed.Footer.Text.ValueString(),
				IconURL:      embed.Footer.IconURL.ValueString(),
				ProxyIconURL: embed.Footer.ProxyIconURL.ValueString(),
			}
		}
		if embed.Provider != nil {
			base.Provider = &discordgo.MessageEmbedProvider{
				Name: embed.Provider.Name.ValueString(),
				URL:  embed.Provider.URL.ValueString(),
			}
		}
		if embed.Author != nil {
			base.Author = &discordgo.MessageEmbedAuthor{
				Name:    embed.Author.Name.ValueString(),
				URL:     embed.Author.URL.ValueString(),
				IconURL: embed.Author.IconURL.ValueString(),
			}
		}
		embeds = append(embeds, base)
	}
	return embeds
}

func unbuildEmbedMessages(embeds []*discordgo.MessageEmbed) []DiscordMessageEmbedModel {
	var ret []DiscordMessageEmbedModel

	for _, embed := range embeds {
		e := DiscordMessageEmbedModel{
			Title:       types.StringValue(embed.Title),
			Description: types.StringValue(embed.Description),
			URL:         types.StringValue(embed.URL),
			Timestamp:   types.StringValue(embed.Timestamp),
			Color:       types.Int64Value(int64(embed.Color)),
		}
		if embed.Image != nil {
			e.Image = &DiscordMessageEmbedImageModel{
				URL:      types.StringValue(embed.Image.URL),
				ProxyURL: types.StringValue(embed.Image.ProxyURL),
				Height:   types.Int64Value(int64(embed.Image.Height)),
				Width:    types.Int64Value(int64(embed.Image.Width)),
			}
		}
		if embed.Thumbnail != nil {
			e.Thumbnail = &DiscordMessageEmbedThumbnailModel{
				URL:      types.StringValue(embed.Thumbnail.URL),
				ProxyURL: types.StringValue(embed.Thumbnail.ProxyURL),
				Height:   types.Int64Value(int64(embed.Thumbnail.Height)),
				Width:    types.Int64Value(int64(embed.Thumbnail.Width)),
			}
		}
		if embed.Video != nil {
			e.Video = &DiscordMessageEmbedVideoModel{
				URL:    types.StringValue(embed.Video.URL),
				Height: types.Int64Value(int64(embed.Video.Height)),
				Width:  types.Int64Value(int64(embed.Video.Width)),
			}
		}
		if embed.Provider != nil {
			e.Provider = &DiscordMessageEmbedProviderModel{
				Name: types.StringValue(embed.Provider.Name),
				URL:  types.StringValue(embed.Provider.URL),
			}
		}
		if embed.Footer != nil {
			e.Footer = &DiscordMessageEmbedFooterModel{
				Text:    types.StringValue(embed.Footer.Text),
				IconURL: types.StringValue(embed.Footer.IconURL),
			}
		}
		if embed.Author != nil {
			e.Author = &DiscordMessageEmbedAuthorModel{
				Name:         types.StringValue(embed.Author.Name),
				URL:          types.StringValue(embed.Author.URL),
				IconURL:      types.StringValue(embed.Author.IconURL),
				ProxyIconURL: types.StringValue(embed.Author.IconURL),
			}
		}
		for _, field := range embed.Fields {
			e.Fields = append(e.Fields, &DiscordMessageEmbedFieldModel{
				Name:   types.StringValue(field.Name),
				Value:  types.StringValue(field.Value),
				Inline: types.BoolValue(field.Inline),
			})
		}
		ret = append(ret, e)
	}
	return ret
}
