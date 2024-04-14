package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/net/context"
)

func dataSourceDiscordMember() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMemberRead,

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_id": {
				ExactlyOneOf: []string{"user_id", "username"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"username": {
				ExactlyOneOf: []string{"user_id", "username"},
				Type:         schema.TypeString,
				Optional:     true,
			},
			"discriminator": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Discriminator is being deprecated by Discord. Only use this if there are users who haven't migrated their username.",
			},
			"joined_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"premium_since": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avatar": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nick": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"in_server": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var member *discordgo.Member
	var memberErr error
	client := m.(*Context).Session
	serverId := d.Get("server_id").(string)

	if v, ok := d.GetOk("user_id"); ok {

		member, memberErr = client.GuildMember(serverId, v.(string), discordgo.WithContext(ctx))
	}

	if v, ok := d.GetOk("username"); ok {
		username := v.(string)
		members, err := client.GuildMembersSearch(serverId, username, 1, discordgo.WithContext(ctx))
		if err != nil {
			return diag.Errorf("Failed to fetch members for %s: %s", serverId, err.Error())
		}

		discriminator := d.Get("discriminator").(string)
		memberErr = fmt.Errorf("failed to find member by name#discriminator: %s#%s", username, discriminator)
		for _, m := range members {
			if m.User.Username == username && m.User.Discriminator == discriminator {
				member = m
				memberErr = nil
				break
			}
		}
	}
	if memberErr != nil {
		return diag.FromErr(memberErr)
	}
	d.Set("in_server", memberErr == nil)
	if memberErr != nil {
		d.Set("joined_at", nil)
		d.Set("premium_since", nil)
		d.Set("roles", nil)
		d.Set("username", nil)
		d.Set("discriminator", nil)
		d.Set("avatar", nil)
		d.Set("nick", nil)
		return diags
	}

	roles := make([]string, 0, len(member.Roles))
	for _, r := range member.Roles {
		roles = append(roles, r)
	}
	if member.PremiumSince == nil {
		d.Set("premium_since", nil)
	}

	d.SetId(member.User.ID)
	d.Set("joined_at", member.JoinedAt.String())
	d.Set("roles", roles)
	d.Set("username", member.User.Username)
	d.Set("discriminator", member.User.Discriminator)
	d.Set("avatar", member.User.Avatar)
	d.Set("nick", member.Nick)

	return diags
}
