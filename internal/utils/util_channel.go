package utils

import (
	"github.com/bwmarrin/discordgo"
)

func GetTextChannelType(channelType discordgo.ChannelType) (string, bool) {
	switch channelType {
	case 0:
		return "text", true
	case 2:
		return "voice", true
	case 4:
		return "category", true
	case 5:
		return "news", true
	case 6:
		return "store", true
	}

	return "text", false
}

type Channel struct {
	ServerId  string
	ChannelId string
	Channel   *discordgo.Channel
}

func GetDiscordChannelPermissionType(value string) (discordgo.PermissionOverwriteType, bool) {
	switch value {
	case "role":
		return discordgo.PermissionOverwriteTypeRole, true
	case "user":
		return discordgo.PermissionOverwriteTypeMember, true
	default:
		return 0, false
	}
}

func GetDiscordChannelPermissionTypeString(value discordgo.PermissionOverwriteType) (string, bool) {
	switch value {
	case discordgo.PermissionOverwriteTypeRole:
		return "role", true
	case discordgo.PermissionOverwriteTypeMember:
		return "user", true
	default:
		return "", false
	}
}
