package discord

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Discord API Token. This can be found in the Discord Developer Portal. This includes the `Bot` prefix. Can also be set via the `DISCORD_TOKEN` environment variable.",
				},
				"client_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"secret": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},

			ResourcesMap: map[string]*schema.Resource{
				"discord_server":             resourceDiscordServer(),
				"discord_managed_server":     resourceDiscordManagedServer(),
				"discord_category_channel":   resourceDiscordCategoryChannel(),
				"discord_text_channel":       resourceDiscordTextChannel(),
				"discord_voice_channel":      resourceDiscordVoiceChannel(),
				"discord_news_channel":       resourceDiscordNewsChannel(),
				"discord_channel_permission": resourceDiscordChannelPermission(),
				"discord_invite":             resourceDiscordInvite(),
				"discord_role":               resourceDiscordRole(),
				"discord_role_everyone":      resourceDiscordRoleEveryone(),
				"discord_member_roles":       resourceDiscordMemberRoles(),
				"discord_message":            resourceDiscordMessage(),
				"discord_system_channel":     resourceDiscordSystemChannel(),
				"discord_webhook":            resourceDiscordWebhook(),
			},

			DataSourcesMap: map[string]*schema.Resource{
				"discord_permission":     dataSourceDiscordPermission(),
				"discord_color":          dataSourceDiscordColor(),
				"discord_local_image":    dataSourceDiscordLocalImage(),
				"discord_role":           dataSourceDiscordRole(),
				"discord_server":         dataSourceDiscordServer(),
				"discord_member":         dataSourceDiscordMember(),
				"discord_system_channel": dataSourceDiscordSystemChannel(),
			},

			ConfigureContextFunc: providerConfigure(version),
		}
		return p
	}
}

func providerConfigure(version string) func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics

		var token string
		if v, ok := d.GetOk("token"); ok {
			token = v.(string)
		} else {
			token = os.Getenv("DISCORD_TOKEN")
		}
		if token == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing required token",
				Detail:   "The `token` argument or `DISCORD_TOKEN` environment variable must be set",
			})
			return nil, diags
		}
		config := Config{
			Token:    token,
			ClientID: d.Get("client_id").(string),
			Secret:   d.Get("secret").(string),
		}

		client, err := config.Client(version)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return client, diags
	}
}
