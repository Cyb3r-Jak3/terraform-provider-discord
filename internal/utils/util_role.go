package utils

import (
	"context"
	"github.com/bwmarrin/discordgo"
)

type Role struct {
	ServerId string
	RoleId   string
	Role     *discordgo.Role
}

func InsertRole(array []*discordgo.Role, value *discordgo.Role, index int) []*discordgo.Role {
	return append(array[:index], append([]*discordgo.Role{value}, array[index:]...)...)
}

func RemoveRole(array []*discordgo.Role, index int) []*discordgo.Role {
	return append(array[:index], array[index+1:]...)
}

func RemoveRoleById(array []string, id string) []string {
	roles := make([]string, 0, len(array))
	for _, x := range array {
		if x != id {
			roles = append(roles, x)
		}
	}

	return roles
}

func MoveRole(array []*discordgo.Role, srcIndex int, dstIndex int) []*discordgo.Role {
	value := array[srcIndex]
	return InsertRole(RemoveRole(array, srcIndex), value, dstIndex)
}

func FindRoleIndex(array []*discordgo.Role, value *discordgo.Role) (int, bool) {
	for index, element := range array {
		if element.ID == value.ID {
			return index, true
		}
	}

	return -1, false
}

func FindRoleById(array []*discordgo.Role, id string) *discordgo.Role {
	for _, element := range array {
		if element.ID == id {
			return element
		}
	}

	return nil
}

//func ReorderRoles(ctx context.Context, s *discordgo.Session, serverId string, role *discordgo.Role, position int) (bool, diag.Diagnostics) {
//
//	roles, err := s.GuildRoles(serverId, discordgo.WithContext(ctx))
//	if err != nil {
//		return false, diag.Errorf("Failed to fetch roles: %s", err.Error())
//	}
//	index, exists := FindRoleIndex(roles, role)
//	if !exists {
//		return false, diag.Errorf("Role somehow does not exists")
//	}
//
//	MoveRole(roles, index, position)
//
//	if roles, err = s.GuildRoleReorder(serverId, roles, discordgo.WithContext(ctx)); err != nil {
//		return false, diag.Errorf("Failed to re-order roles: %s", err.Error())
//	}
//
//	return true, nil
//}

func GetRole(ctx context.Context, client *discordgo.Session, serverId string, roleId string) (*discordgo.Role, error) {
	if roles, err := client.GuildRoles(serverId, discordgo.WithContext(ctx)); err != nil {
		return nil, err
	} else {
		return FindRoleById(roles, roleId), nil
	}
}
