package discord

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestAccDatasourceDiscordRole(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	testRoleID := os.Getenv("DISCORD_TEST_ROLE_ID")
	testRoleName := os.Getenv("DISCORD_TEST_ROLE_NAME")
	if testServerID == "" || testRoleID == "" || testRoleName == "" {
		t.Skip("DISCORD_TEST_SERVER_ID, DISCORD_TEST_ROLE_ID, and DISCORD_TEST_ROLE_NAME envvars must be set for acceptance tests")
	}

	name := "data.discord_role.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordRoleID(testServerID, testRoleID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "server_id", testServerID),
					resource.TestCheckResourceAttr(name, "role_id", testRoleID),
					resource.TestCheckResourceAttr(name, "position", "1"),
					resource.TestCheckResourceAttrSet(name, "permissions"),
					resource.TestCheckResourceAttrSet(name, "color"),
					resource.TestCheckResourceAttr(name, "hoist", "false"),
					resource.TestCheckResourceAttr(name, "mentionable", "false"),
					resource.TestCheckResourceAttr(name, "managed", "false"),
				),
			},
			{
				Config: testAccDatasourceDiscordRoleName(testServerID, testRoleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "server_id", testServerID),
					resource.TestCheckResourceAttr(name, "name", testRoleName),
					resource.TestCheckResourceAttr(name, "position", "1"),
					resource.TestCheckResourceAttrSet(name, "permissions"),
					resource.TestCheckResourceAttrSet(name, "color"),
					resource.TestCheckResourceAttr(name, "hoist", "false"),
					resource.TestCheckResourceAttr(name, "mentionable", "false"),
					resource.TestCheckResourceAttr(name, "managed", "false"),
				),
			},
		},
	})
}

func testAccDatasourceDiscordRoleID(serverId string, roleID string) string {
	return fmt.Sprintf(`
	data "discord_role" "example" {
	  server_id = "%[1]s"
      role_id = "%[2]s"
	}`, serverId, roleID)
}

func testAccDatasourceDiscordRoleName(serverId string, roleName string) string {
	return fmt.Sprintf(`
	data "discord_role" "example" {
	  server_id = "%[1]s"
	  name = "%[2]s"
	}`, serverId, roleName)
}
