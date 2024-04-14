package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/polds/imgbase64"
	"golang.org/x/net/context"
)

func baseServerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"verification_level": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			ValidateFunc: func(val interface{}, key string) (warns []string, errors []error) {
				v := val.(int)
				if v > 4 || v < 0 {
					errors = append(errors, fmt.Errorf("verification_level must be between 0 and 4 inclusive, got: %d", v))
				}

				return
			},
		},
		"explicit_content_filter": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			ValidateFunc: func(val interface{}, key string) (warns []string, errors []error) {
				v := val.(int)
				if v > 2 || v < 0 {
					errors = append(errors, fmt.Errorf("explicit_content_filter must be between 0 and 2 inclusive, got: %d", v))
				}

				return
			},
		},
		"default_message_notifications": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			ValidateFunc: func(val interface{}, key string) (warns []string, errors []error) {
				v := val.(int)
				if v != 0 && v != 1 {
					errors = append(errors, fmt.Errorf("default_message_notifications must be 0 or 1, got: %d", v))
				}

				return
			},
		},
		"afk_channel_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"afk_timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  300,
			ValidateFunc: func(val interface{}, key string) (warns []string, errors []error) {
				v := val.(int)
				// See: https://discord.com/developers/docs/resources/guild#guild-object-guild-structure
				expected := []int{60, 300, 900, 1800, 3600}
				if !contains(expected, v) {
					errors = append(errors, fmt.Errorf("afk_timeout must be set to one of the following values: %d, but got: %d", expected, v))
				}

				return
			},
		},
		"icon_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"icon_data_uri": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"icon_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"splash_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"splash_data_uri": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"splash_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"owner_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
	}
}

func managedServerSchema() map[string]*schema.Schema {
	res := baseServerSchema()

	res["server_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
	res["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}

	return res
}

func serverSchema() map[string]*schema.Schema {
	res := baseServerSchema()

	res["server_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	res["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}

	return res
}

func resourceDiscordServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: serverSchema(),
	}
}

func resourceDiscordManagedServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerManagedCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerManagedDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: managedServerSchema(),
	}
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	icon := ""
	if v, ok := d.GetOk("icon_url"); ok {
		icon = imgbase64.FromRemote(v.(string))
	}
	if v, ok := d.GetOk("icon_data_uri"); ok {
		icon = v.(string)
	}

	name := d.Get("name").(string)
	// DiscordGo doesn't support creating a server with anything apart from a name.
	server, err := client.GuildCreate(name, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to create server: %s", err.Error())
	}

	splash := ""
	if v, ok := d.GetOk("splash_url"); ok {
		splash = imgbase64.FromRemote(v.(string))
	}
	if v, ok := d.GetOk("splash_data_uri"); ok {
		splash = v.(string)
	}
	if splash != "" {
	}

	afkChannel := server.AfkChannelID
	if v, ok := d.GetOk("afk_channel_id"); ok {
		afkChannel = v.(string)
	}
	afkTimeOut := server.AfkTimeout
	if v, ok := d.GetOk("afk_timeout"); ok {
		// The value has been already validated, so this cast is safe.
		afkTimeOut = v.(int)
	}

	verificationLevel := discordgo.VerificationLevel(d.Get("verification_level").(int))
	server, err = client.GuildEdit(server.ID, &discordgo.GuildParams{
		Icon:                        icon,
		Region:                      d.Get("region").(string),
		VerificationLevel:           &verificationLevel,
		DefaultMessageNotifications: d.Get("default_message_notifications").(int),
		ExplicitContentFilter:       d.Get("explicit_content_filter").(int),
		AfkChannelID:                afkChannel,
		AfkTimeout:                  afkTimeOut,
		Splash:                      splash,
	}, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to edit server: %s", err.Error())
	}

	for _, channel := range server.Channels {
		if _, err := client.ChannelDelete(channel.ID); err != nil {
			return diag.Errorf("Failed to delete channel for new server: %s", err.Error())
		}
	}

	// Update owner's ID if the specified one is not as same as default,
	// because we will receive "User is already owner" error if update to the same one.
	ownerId := server.OwnerID
	if v, ok := d.GetOk("owner_id"); ok {
		newOwnerId := v.(string)
		if ownerId != newOwnerId {
			ownerId = newOwnerId
		}
		server, err = client.GuildEdit(server.ID, &discordgo.GuildParams{
			OwnerID: ownerId,
		}, discordgo.WithContext(ctx))
	}

	d.SetId(server.ID)
	if _, ok := d.GetOk("owner_id"); !ok {
		d.Set("owner_id", server.OwnerID)
	}
	d.Set("region", server.Region)
	d.Set("icon_hash", server.Icon)
	d.Set("splash_hash", server.Splash)

	return diags
}

func resourceServerManagedCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	serverIdInterface, ok := d.GetOk("server_id")
	if !ok {
		return diag.Errorf("Error: server_id must be set")
	}
	serverId, ok := serverIdInterface.(string)
	if !ok {
		return diag.Errorf("Error: server_id must be a string")
	}

	d.SetId(serverId)

	return diags
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	server, err := client.Guild(d.Id(), discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Error fetching server: %s", err.Error())
	}

	d.Set("name", server.Name)
	d.Set("region", server.Region)
	d.Set("default_message_notifications", server.DefaultMessageNotifications)
	d.Set("afk_timeout", server.AfkTimeout)
	d.Set("icon_hash", server.Icon)
	d.Set("splash_hash", server.Splash)
	d.Set("verification_level", server.VerificationLevel)
	d.Set("default_message_notifications", server.DefaultMessageNotifications)
	d.Set("explicit_content_filter", server.ExplicitContentFilter)
	if server.AfkChannelID != "" {
		d.Set("afk_channel_id", server.AfkChannelID)
	}

	// We don't want to set the owner to null, should only change this if it's changing to something else
	if d.Get("owner_id").(string) != "" && server.OwnerID != "" {
		d.Set("owner_id", server.OwnerID)
	}

	return diags
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	server, err := client.Guild(d.Id(), discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Error fetching server: %s", err.Error())
	}

	guildParams := &discordgo.GuildParams{}
	edit := false

	if d.HasChange("icon_url") {
		guildParams.Icon = imgbase64.FromRemote(d.Get("icon_url").(string))
		edit = true
	}
	if d.HasChange("icon_data_uri") {
		guildParams.Icon = d.Get("icon_data_uri").(string)
		edit = true
	}
	if d.HasChange("splash_url") {
		guildParams.Splash = imgbase64.FromRemote(d.Get("splash_url").(string))
		edit = true
	}
	if d.HasChange("splash_data_uri") {
		guildParams.Splash = d.Get("splash_data_uri").(string)
		edit = true
	}
	if d.HasChange("afk_channel_id") {
		guildParams.AfkChannelID = d.Get("afk_channel_id").(string)
		edit = true
	}
	if d.HasChange("afk_timeout") {
		guildParams.AfkTimeout = d.Get("afk_timeout").(int)
		edit = true
	}

	if d.HasChange("owner_id") {
		guildParams.OwnerID = d.Get("owner_id").(string)
		edit = true
	}
	if d.HasChange("verification_level") {
		verificationLevel := discordgo.VerificationLevel(d.Get("verification_level").(int))
		guildParams.VerificationLevel = &verificationLevel
		edit = true
	}

	if d.HasChange("default_message_notifications") {
		guildParams.DefaultMessageNotifications = d.Get("default_message_notifications").(int)
		edit = true
	}
	if d.HasChange("explicit_content_filter") {
		guildParams.ExplicitContentFilter = d.Get("explicit_content_filter").(int)
		edit = true
	}
	if d.HasChange("name") {
		guildParams.Name = d.Get("name").(string)
		edit = true
	}
	if d.HasChange("region") {
		guildParams.Region = d.Get("region").(string)
		edit = true
	}

	ownerId, hasOwner := d.GetOk("owner_id")
	if d.HasChange("owner_id") {
		if hasOwner {
			guildParams.OwnerID = ownerId.(string)
			edit = true
		}
	} else {
		if hasOwner {
			guildParams.OwnerID = server.OwnerID
			edit = true
		}
	}

	if edit {
		if _, err = client.GuildEdit(server.ID, guildParams, discordgo.WithContext(ctx)); err != nil {
			return diag.Errorf("Failed to edit server: %s", err.Error())
		}
	}

	return diags
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	if err := client.GuildDelete(d.Id(), discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to delete server: %s", err)
	}

	return diags
}

func resourceServerManagedDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// noop

	return diags
}
