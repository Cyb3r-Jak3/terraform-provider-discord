package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func resourceDiscordRoleEveryone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleEveryoneRead,
		ReadContext:   resourceRoleEveryoneRead,
		UpdateContext: resourceRoleEveryoneUpdate,
		DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
			return []diag.Diagnostic{{
				Severity: diag.Warning,
				Summary:  "Deleting the everyone role is not allowed",
			}}
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceRoleEveryoneImport,
		},

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permissions": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ForceNew: false,
			},
		},
	}
}

func resourceRoleEveryoneImport(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	data.SetId(data.Id())
	data.Set("server_id", data.Id())

	return schema.ImportStatePassthroughContext(ctx, data, i)
}

func resourceRoleEveryoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	d.SetId(serverId)

	if role, err := getRole(ctx, client, serverId, serverId); err != nil {
		return diag.Errorf("Failed to fetch role %s: %s", d.Id(), err.Error())
	} else {
		d.Set("permissions", role.Permissions)

		return diags
	}
}

func resourceRoleEveryoneUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	d.SetId(serverId)
	newPermission := int64(d.Get("permissions").(int))

	if role, err := client.GuildRoleEdit(serverId, serverId, &discordgo.RoleParams{
		Permissions: &newPermission,
	}, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to update role %s: %s", d.Id(), err.Error())
	} else {
		d.Set("permissions", role.Permissions)

		return diags
	}
}
