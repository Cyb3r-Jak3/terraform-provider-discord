package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"

	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordChannelPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChannelPermissionCreate,
		ReadContext:   resourceChannelPermissionRead,
		UpdateContext: resourceChannelPermissionUpdate,
		DeleteContext: resourceChannelPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"role", "user"}, false),
			},
			"overwrite_id": {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeString,
			},
			"allow": {
				AtLeastOneOf: []string{"allow", "deny"},
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"deny": {
				AtLeastOneOf: []string{"allow", "deny"},
				Optional:     true,
				Type:         schema.TypeInt,
			},
		},
	}
}

func resourceChannelPermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	overwriteId := d.Get("overwrite_id").(string)
	permissionType, _ := getDiscordChannelPermissionType(d.Get("type").(string))
	if err := client.ChannelPermissionSet(
		channelId, overwriteId, permissionType,
		int64(d.Get("allow").(int)),
		int64(d.Get("deny").(int)), discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to update channel permissions %s: %s", channelId, err.Error())
	} else {
		d.SetId(generateThreePartId(channelId, overwriteId, d.Get("type").(string)))

		return diags
	}
}

func resourceChannelPermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	overwriteId := d.Get("overwrite_id").(string)
	var permissionType discordgo.PermissionOverwriteType

	cId, oId, pt, err := parseThreeIds(d.Id())
	if err != nil {
		log.Default().Printf("Unable to parse IDs out of the resource ID. Falling back on legacy config behavior.")
		channelId = d.Get("channel_id").(string)
		overwriteId = d.Get("overwrite_id").(string)
		permissionType, _ = getDiscordChannelPermissionType(d.Get("type").(string))
	} else {
		channelId = cId
		overwriteId = oId
		permissionType, _ = getDiscordChannelPermissionType(pt)

		d.Set("channel_id", channelId)
		d.Set("overwrite_id", overwriteId)
		d.Set("type", pt)
	}

	channel, err := client.Channel(channelId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to find channel %s: %s", channelId, err.Error())
	}

	for _, x := range channel.PermissionOverwrites {
		if uint(x.Type) == uint(permissionType) && x.ID == overwriteId {
			d.Set("allow", int(x.Allow))
			d.Set("deny", int(x.Deny))
			break
		}
	}

	return diags
}

func resourceChannelPermissionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	overwriteId := d.Get("overwrite_id").(string)
	permissionType, _ := getDiscordChannelPermissionType(d.Get("type").(string))

	if err := client.ChannelPermissionSet(
		channelId, overwriteId, permissionType,
		d.Get("allow").(int64),
		d.Get("deny").(int64), discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to update channel permissions %s: %s", channelId, err.Error())
	} else {
		d.SetId(strconv.Itoa(
			Hashcode(
				fmt.Sprintf(
					"%s:%s:%s", channelId, overwriteId, d.Get("type").(string),
				),
			),
		),
		)
		return diags
	}
}

func resourceChannelPermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	channelId := d.Get("channel_id").(string)
	overwriteId := d.Get("overwrite_id").(string)

	if err := client.ChannelPermissionDelete(channelId, overwriteId, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to delete channel permissions %s: %s", channelId, err.Error())
	} else {
		return diags
	}
}
