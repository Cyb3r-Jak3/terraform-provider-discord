package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordChannelPermission(t *testing.T) {
	testChannelID := os.Getenv("DISCORD_TEST_CHANNEL_ID")
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	testRoleID := os.Getenv("DISCORD_TEST_ROLE_ID")
	if testChannelID == "" || testServerID == "" || testRoleID == "" {
		t.Skip("DISCORD_TEST_CHANNEL_ID envvar must be set for acceptance tests")
	}
	name := "discord_channel_permission.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordChannelPermission(testServerID, testChannelID, testRoleID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "channel_id", testChannelID),
					resource.TestCheckResourceAttr(name, "type", "role"),
					resource.TestCheckResourceAttr(name, "overwrite_id", testRoleID),
					resource.TestCheckResourceAttr(name, "allow", "1024"),
				),
			},
			//{
			//	ResourceName:        name,
			//	ImportState:         true,
			//	ImportStateVerify:   true,
			//	ImportStateIdPrefix: fmt.Sprintf("%s/%s/", testChannelID, "role"),
			//},
		},
	})
}

func testAccResourceDiscordChannelPermission(serverID, channelID, roleID string) string {
	return fmt.Sprintf(`
    data "discord_role" "example" {
	  server_id = "%[1]s"
      id = "%[2]s"
	}
	resource "discord_channel_permission" "example" {
      channel_id = "%[3]s"
	  type = "role"
      overwrite_id = data.discord_role.example.id
      allow = 1024
	}`, serverID, roleID, channelID)
}
