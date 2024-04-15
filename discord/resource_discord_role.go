package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/net/context"
)

func resourceDiscordRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRoleImport,
		},

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"permissions": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ForceNew: false,
			},
			"color": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},
			"hoist": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: false,
			},
			"mentionable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: false,
			},
			"position": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ForceNew:     false,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceRoleImport(ctx context.Context, data *schema.ResourceData, i interface{}) ([]*schema.ResourceData, error) {
	if serverId, roleId, err := parseTwoIds(data.Id()); err != nil {
		return nil, err
	} else {
		data.SetId(roleId)
		data.Set("server_id", serverId)

		return schema.ImportStatePassthroughContext(ctx, data, i)
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	server, err := client.Guild(serverId)
	if err != nil {
		return diag.Errorf("Server does not exist with that ID: %s", serverId)
	}
	role, err := client.GuildRoleCreate(serverId, &discordgo.RoleParams{
		Name:        d.Get("name").(string),
		Permissions: Int64Ptr(int64(d.Get("permissions").(int))),
		Color:       IntPtr(d.Get("color").(int)),
		Hoist:       BoolPtr(d.Get("hoist").(bool)),
		Mentionable: BoolPtr(d.Get("mentionable").(bool)),
	}, discordgo.WithContext(ctx))

	if err != nil {
		return diag.Errorf("Failed to create role for %s: %s", serverId, err.Error())
	}

	if newPosition, ok := d.GetOk("position"); ok {
		var oldRole *discordgo.Role
		for _, r := range server.Roles {
			if r.Position == newPosition.(int) {
				oldRole = r
				break
			}
		}

		param := []*discordgo.Role{{ID: role.ID, Position: newPosition.(int)}}
		if oldRole != nil {
			param = append(param, &discordgo.Role{ID: oldRole.ID, Position: role.Position})
		}

		if _, err := client.GuildRoleReorder(serverId, param, discordgo.WithContext(ctx)); err != nil {
			diags = append(diags, diag.Errorf("Failed to re-order roles: %s", err.Error())...)
		} else {
			d.Set("position", newPosition)
		}
	}

	d.SetId(role.ID)
	d.Set("server_id", server.ID)
	d.Set("managed", role.Managed)

	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	role, err := getRole(ctx, client, d.Get("server_id").(string), d.Id())

	if err != nil {
		return diag.Errorf("Failed to fetch role %s: %s", d.Id(), err.Error())
	}

	d.Set("name", role.Name)
	d.Set("position", role.Position)
	d.Set("color", role.Color)
	d.Set("hoist", role.Hoist)
	d.Set("mentionable", role.Mentionable)
	d.Set("permissions", role.Permissions)
	d.Set("managed", role.Managed)

	return diags

}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id").(string)
	server, err := client.Guild(serverId)
	if err != nil {
		return diag.Errorf("Failed to fetch server %s: %s", serverId, err.Error())
	}

	roleId := d.Id()
	role, err := getRole(ctx, client, serverId, roleId)
	if err != nil {
		return diag.Errorf("Failed to fetch role %s: %s", d.Id(), err.Error())
	}

	if d.HasChange("position") {
		_, newPosition := d.GetChange("position")
		var oldRole *discordgo.Role
		for _, r := range server.Roles {
			if r.Position == newPosition.(int) {
				oldRole = r
				break
			}
		}

		param := []*discordgo.Role{{ID: role.ID, Position: newPosition.(int)}}
		if oldRole != nil {
			param = append(param, &discordgo.Role{ID: oldRole.ID, Position: role.Position})
		}

		if _, err := client.GuildRoleReorder(serverId, param, discordgo.WithContext(ctx)); err != nil {
			diags = append(diags, diag.Errorf("Failed to re-order roles: %s", err.Error())...)
		} else {
			d.Set("position", newPosition)
		}
	}

	var (
		newName        = d.Get("name").(string)
		newColor       int
		newHoist       = d.Get("hoist").(bool)
		newMentionable = d.Get("mentionable").(bool)
		newPermissions = d.Get("permissions").(int64)
	)
	if _, v := d.GetChange("color"); v.(int) > 0 {
		newColor = v.(int)
	} else {
		newColor = role.Color
	}

	if role, err = client.GuildRoleEdit(serverId, roleId, &discordgo.RoleParams{
		Name:        newName,
		Color:       &newColor,
		Hoist:       &newHoist,
		Mentionable: &newMentionable,
		Permissions: &newPermissions,
	}, discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to update role %s: %s", d.Id(), err.Error())
	}

	d.Set("name", role.Name)
	d.Set("position", role.Position)
	d.Set("color", role.Color)
	d.Set("hoist", role.Hoist)
	d.Set("mentionable", role.Mentionable)
	d.Set("permissions", role.Permissions)
	d.Set("managed", role.Managed)

	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*Context).Session

	serverId := d.Get("server_id")
	if err := client.GuildRoleDelete(serverId.(string), d.Id(), discordgo.WithContext(ctx)); err != nil {
		return diag.Errorf("Failed to delete role: %s", err.Error())
	}

	return diags
}
