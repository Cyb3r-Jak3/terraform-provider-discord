package discord

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordVoiceChannel(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}
	name := "discord_voice_channel.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordVoiceChannel(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "server_id", testServerID),
					resource.TestCheckResourceAttr(name, "name", "terraform-voice-channel"),
					resource.TestCheckResourceAttr(name, "type", "voice"),
					resource.TestCheckResourceAttr(name, "position", "1"),
					resource.TestCheckResourceAttr(name, "bitrate", "64000"),
					resource.TestCheckResourceAttr(name, "user_limit", "4"),
					resource.TestCheckResourceAttr(name, "sync_perms_with_category", "false"),
					resource.TestCheckResourceAttrSet(name, "channel_id"),
				),
			},
		},
	})
}

func testAccResourceDiscordVoiceChannel(serverID string) string {
	return fmt.Sprintf(`
	resource "discord_voice_channel" "example" {
	  server_id = "%[1]s"
      name = "terraform-voice-channel"
      position = 1
      bitrate = 64000
      user_limit = 4
      sync_perms_with_category = false
	}`, serverID)
}
