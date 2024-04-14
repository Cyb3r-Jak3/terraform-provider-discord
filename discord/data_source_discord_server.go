package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDiscordServer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDiscordServerRead,
		Schema: map[string]*schema.Schema{
			"server_id": {
				ExactlyOneOf: []string{"server_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"name": {
				ExactlyOneOf: []string{"server_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_message_notifications": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"verification_level": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"explicit_content_filter": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"afk_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"icon_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"splash_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"afk_channel_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDiscordServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var server *discordgo.Guild
	client := m.(*Context).Session

	if v, ok := d.GetOk("server_id"); ok {
		server, err = client.Guild(v.(string), discordgo.WithContext(ctx))
		if err != nil {
			return diag.Errorf("Failed to fetch server %s: %s", v.(string), err.Error())
		}
	}
	if v, ok := d.GetOk("name"); ok {
		guilds, err := client.UserGuilds(1000, "", "", false, discordgo.WithContext(ctx))
		if err != nil {
			return diag.Errorf("Failed to fetch server %s: %s", v.(string), err.Error())
		}

		for _, s := range guilds {
			if s.Name == v.(string) {
				server, err = client.Guild(v.(string), discordgo.WithContext(ctx))
				if err != nil {
					return diag.Errorf("Failed to fetch server %s: %s", v.(string), err.Error())
				}
				break
			}
		}

		if server == nil {
			return diag.Errorf("Failed to fetch server %s", v.(string))
		}
	}

	d.SetId(server.ID)
	d.Set("server_id", server.ID)
	d.Set("name", server.Name)
	d.Set("region", server.Region)
	d.Set("afk_timeout", server.AfkTimeout)
	d.Set("icon_hash", server.Icon)
	d.Set("splash_hash", server.Splash)
	d.Set("default_message_notifications", int(server.DefaultMessageNotifications))
	d.Set("verification_level", int(server.VerificationLevel))
	d.Set("explicit_content_filter", int(server.ExplicitContentFilter))

	if server.AfkChannelID != "" {
		d.Set("afk_channel_id", server.AfkChannelID)
	}
	if server.OwnerID != "" {
		d.Set("owner_id", server.OwnerID)
	}

	return diags
}
