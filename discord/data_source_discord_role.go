package discord

import (
	"context"
	"github.com/bwmarrin/discordgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDiscordRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDiscordRoleRead,
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_id": {
				ExactlyOneOf: []string{"role_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"name": {
				ExactlyOneOf: []string{"role_id", "name"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"position": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"color": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"permissions": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hoist": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mentionable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDiscordRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var role *discordgo.Role
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	server, err := client.Guild(serverId, discordgo.WithContext(ctx))
	if err != nil {
		return diag.Errorf("Failed to fetch server %s: %s", serverId, err.Error())
	}

	roleID := d.Get("role_id").(string)
	roleName := d.Get("name").(string)
	for _, r := range server.Roles {
		if r.ID == roleID || r.Name == roleName {
			role = r
			break
		}
	}
	if role == nil {
		return diag.Errorf("Failed to find role by ID %s or name: %s", roleID, roleName)
	}

	d.SetId(role.ID)
	d.Set("role_id", role.ID)
	d.Set("name", role.Name)
	d.Set("position", role.Position)
	d.Set("color", role.Color)
	d.Set("hoist", role.Hoist)
	d.Set("mentionable", role.Mentionable)
	d.Set("permissions", int(role.Permissions))
	d.Set("managed", role.Managed)

	return diags
}
