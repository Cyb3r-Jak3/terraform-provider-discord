package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordInvite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInviteCreate,
		ReadContext:   resourceInviteRead,
		DeleteContext: resourceInviteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"max_age": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
				Default:  86400,
			},
			"max_uses": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"temporary": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},
			"unique": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},
			"code": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceInviteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)

	if invite, err := client.ChannelInviteCreate(channelId, discordgo.Invite{
		MaxAge:    d.Get("max_age").(int),
		MaxUses:   d.Get("max_uses").(int),
		Temporary: d.Get("temporary").(bool),
		Unique:    d.Get("unique").(bool),
	}, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to create a invite: %s", err.Error())
	} else {
		d.SetId(invite.Code)
		d.Set("code", invite.Code)

		return diags
	}
}

func resourceInviteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	if invite, err := client.Invite(d.Id(), discordgo.WithContext(ctx)); err != nil {
		d.SetId("")
	} else {
		d.Set("code", invite.Code)
	}

	return diags
}

func resourceInviteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	if _, err := client.InviteDelete(d.Id(), discordgo.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	} else {
		return diags
	}
}
