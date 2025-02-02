package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordChannelCategory(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}
	name := "discord_category_channel.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordSystemChannel(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "server_id", testServerID),
					resource.TestCheckResourceAttr(name, "name", "terraform-system-channel"),
					resource.TestCheckResourceAttr(name, "type", "category"),
					resource.TestCheckResourceAttr(name, "position", "1"),
					resource.TestCheckResourceAttrSet(name, "channel_id"),
					resource.TestCheckResourceAttr("discord_text_channel.example", "sync_perms_with_category", "true"),
					resource.TestCheckResourceAttr("discord_text_channel.example2", "sync_perms_with_category", "false"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceDiscordSystemChannel(serverID string) string {
	return fmt.Sprintf(`
	resource "discord_category_channel" "example" {
	  server_id = "%[1]s"
      name = "terraform-system-channel"
      position = 1
	}

	resource "discord_text_channel" "example" {
	  server_id = "%[1]s"
      name = "terraform-text-channel"
      position = 2
      topic = "Testing text channel in category"
      sync_perms_with_category = true
      category = discord_category_channel.example.channel_id
	}

	resource "discord_text_channel" "example2" {
	  server_id = "%[1]s"
      name = "terraform-text-channel-2"
      position = 3
      topic = "Testing text channel 2 in category"
      sync_perms_with_category = false
      category = discord_category_channel.example.channel_id
	}

`, serverID)
}
