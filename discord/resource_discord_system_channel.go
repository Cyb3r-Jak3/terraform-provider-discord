package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordSystemChannel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemChannelCreate,
		ReadContext:   resourceSystemChannelRead,
		UpdateContext: resourceSystemChannelUpdate,
		DeleteContext: resourceSystemChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"system_channel_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSystemChannelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)

	server, err := client.Guild(serverId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to find server: %s", err.Error())
	}

	systemChannelId := server.SystemChannelID
	if v, ok := d.GetOk("system_channel_id"); ok {
		systemChannelId = v.(string)

	} else {
		return diag.Errorf("Failed to parse system channel id")
	}
	if _, err := client.GuildEdit(serverId, &discordgo.GuildParams{
		SystemChannelID: systemChannelId,
	}, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to edit server: %s", err.Error())
	}

	d.SetId(serverId)

	return diags
}

func resourceSystemChannelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Id()

	server, err := client.Guild(serverId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Error fetching server: %s", err.Error())
	}

	d.Set("system_channel_id", server.SystemChannelID)

	return diags
}

func resourceSystemChannelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	_, err := client.Guild(serverId, discordgo.WithContext(ctx))

	if err != nil {
		return diag.Errorf("Error fetching server: %s", err.Error())
	}

	if d.HasChange("system_channel_id") {
		id := d.Get("system_channel_id").(string)
		if _, err := client.GuildEdit(serverId, &discordgo.GuildParams{
			SystemChannelID: id,
		}, discordgo.WithContext(ctx)); err != nil {
			return diag.Errorf("Failed to edit server: %s", err.Error())
		}
	}

	return diags
}

func resourceSystemChannelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverID := d.Get("server_id").(string)

	if _, err := client.GuildEdit(serverID, &discordgo.GuildParams{
		SystemChannelID: "",
	}, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to edit server: %s: %s", serverID, err.Error())
	}

	return diags
}
