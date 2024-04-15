package discord

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordTextChannel(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}
	name := "discord_text_channel.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordTextChannel(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "server_id", testServerID),
					resource.TestCheckResourceAttr(name, "name", "terraform-text-channel"),
					resource.TestCheckResourceAttr(name, "type", "text"),
					resource.TestCheckResourceAttr(name, "position", "1"),
					resource.TestCheckResourceAttrSet(name, "channel_id"),
					resource.TestCheckResourceAttr(name, "topic", "Testing text channel"),
					resource.TestCheckResourceAttr(name, "nsfw", "false"),
					resource.TestCheckResourceAttr(name, "sync_perms_with_category", "false"),
				),
			},
		},
	})
}

func testAccResourceDiscordTextChannel(serverID string) string {
	return fmt.Sprintf(`
	resource "discord_text_channel" "example" {
	  server_id = "%[1]s"
      name = "terraform-text-channel"
      position = 1
      topic = "Testing text channel"
      nsfw = false
      sync_perms_with_category = false
	}`, serverID)
}
