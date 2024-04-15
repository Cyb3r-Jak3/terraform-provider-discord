package discord

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordMessage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMessageCreate,
		ReadContext:   resourceMessageRead,
		UpdateContext: resourceMessageUpdate,
		DeleteContext: resourceMessageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"author": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				AtLeastOneOf: []string{"content", "embed"},
				Type:         schema.TypeString,
				Optional:     true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == strings.TrimSuffix(new, "\r\n")
				},
			},
			"timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"edited_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"tts": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"embed": {
				AtLeastOneOf: []string{"content", "embed"},
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"title": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"timestamp": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"color": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"footer": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text": {
										Type:     schema.TypeString,
										Required: true,
									},
									"icon_url": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"image": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:     schema.TypeString,
										Required: true,
									},
									"proxy_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"height": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"width": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"thumbnail": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:     schema.TypeString,
										Required: true,
									},
									"proxy_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"height": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"width": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"video": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Type:     schema.TypeString,
										Required: true,
									},
									"height": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"width": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"provider": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"url": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"author": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"url": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"icon_url": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"proxy_icon_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"fields": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"inline": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"pinned": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceMessageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)

	embeds := make([]*discordgo.MessageEmbed, 0, 1)
	if v, ok := d.GetOk("embed"); ok {
		if embed, err := buildEmbed(v.([]interface{})); err != nil {
			return diag.Errorf("Failed to create message in %s: %s", channelId, err.Error())
		} else {
			embeds = append(embeds, embed)
		}
	}
	message, err := client.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
		Content: d.Get("content").(string),
		Embeds:  embeds,
		TTS:     d.Get("tts").(bool),
	}, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to create message in %s: %s", channelId, err.Error())
	}

	d.SetId(message.ID)
	d.Set("type", int(message.Type))
	d.Set("timestamp", message.Timestamp.Format(time.RFC3339))
	d.Set("author", message.Author.ID)
	if len(message.Embeds) > 0 {
		d.Set("embed", unbuildEmbed(message.Embeds[0]))
	} else {
		d.Set("embed", nil)
	}
	if message.GuildID != "" {
		d.Set("server_id", message.GuildID)
	}

	if d.Get("pinned").(bool) {
		pinError := client.ChannelMessagePin(channelId, message.ID, discordgo.WithContext(ctx))
		if pinError != nil {
			diags = append(diags, diag.Errorf("Failed to pin message %s in %s: %s", message.ID, channelId, err.Error())...)
		}
	}

	return diags
}

func resourceMessageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	messageId := d.Id()
	message, err := client.ChannelMessage(channelId, messageId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to fetch message %s in %s: %s", messageId, channelId, err.Error())
	}

	if message.GuildID != "" {
		d.Set("server_id", message.GuildID)
	}
	d.Set("type", int(message.Type))
	d.Set("tts", message.TTS)
	d.Set("timestamp", message.Timestamp.Format(time.RFC3339))
	d.Set("author", message.Author.ID)
	d.Set("content", message.Content)
	d.Set("pinned", message.Pinned)

	if len(message.Embeds) > 0 {
		d.Set("embed", unbuildEmbed(message.Embeds[0]))
	}
	if message.EditedTimestamp != nil {
		d.Set("edited_timestamp", message.EditedTimestamp.Format(time.RFC3339))
	}

	return diags
}

func resourceMessageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	messageId := d.Id()

	var content string
	message, err := client.ChannelMessage(channelId, messageId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to fetch message %s in %s: %s", messageId, channelId, err.Error())
	}
	if d.HasChange("content") {
		content = d.Get("content").(string)
	} else {
		content = message.Content
	}

	embeds := make([]*discordgo.MessageEmbed, 0, 1)
	if d.HasChange("embed") {
		var embed *discordgo.MessageEmbed
		_, n := d.GetChange("embed")
		if len(n.([]interface{})) > 0 {
			if e, err := buildEmbed(n.([]interface{})); err != nil {
				return diag.Errorf("Failed to edit message %s in %s: %s", messageId, channelId, err.Error())
			} else {
				embed = e
			}
		}

		embeds = append(embeds, embed)
	}

	editedMessage, err := client.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      messageId,
		Channel: channelId,
		Content: &content,
		Embeds:  &embeds,
	}, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to update message %s in %s: %s", channelId, messageId, err.Error())
	}

	if len(editedMessage.Embeds) > 0 {
		d.Set("embed", unbuildEmbed(message.Embeds[0]))
	} else {
		d.Set("embed", nil)
	}

	d.Set("edited_timestamp", editedMessage.EditedTimestamp.Format(time.RFC3339))

	return diags
}

func resourceMessageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	messageId := d.Id()
	if err := client.ChannelMessageDelete(channelId, messageId, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to delete message %s in %s: %s", messageId, channelId, err.Error())
	} else {
		return diags
	}
}
