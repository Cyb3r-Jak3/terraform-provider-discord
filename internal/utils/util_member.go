package utils

import (
	"github.com/bwmarrin/discordgo"
)

func HasRole(member *discordgo.Member, roleId string) bool {
	for _, r := range member.Roles {
		if r == roleId {
			return true
		}
	}

	return false
}
