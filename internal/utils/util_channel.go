package utils

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
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
	case 7:
		return "forum", true
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

func ParseThreeIds(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected attribute1/attribute2/attriburte3", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func GenerateThreePartId(one string, two string, three string) string {
	return fmt.Sprintf("%s/%s/%s", one, two, three)
}
