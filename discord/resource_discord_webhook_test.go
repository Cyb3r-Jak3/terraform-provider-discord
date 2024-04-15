package discord

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordWebhook(t *testing.T) {
	testChannelID := os.Getenv("DISCORD_TEST_CHANNEL_ID")
	if testChannelID == "" {
		t.Skip("DISCORD_TEST_CHANNEL_ID envvar must be set for acceptance tests")
	}
	name := "discord_webhook.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordWebhook(testChannelID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "channel_id", testChannelID),
					resource.TestCheckResourceAttr(name, "name", "terraform-test"),
					resource.TestCheckResourceAttr(name, "avatar_url", "https://www.terraform.io/assets/images/og-image-8b3e4f7d.png"),
					resource.TestCheckResourceAttrSet(name, "token"),
					resource.TestCheckResourceAttrSet(name, "url"),
					resource.TestCheckResourceAttrSet(name, "slack_url"),
					resource.TestCheckResourceAttrSet(name, "github_url"),
				),
			},
		},
	})
}

func testAccResourceDiscordWebhook(channelID string) string {
	return fmt.Sprintf(`
	resource "discord_webhook" "example" {
      channel_id = "%[1]s"
      name = "terraform-test"
	  avatar_url = "https://www.terraform.io/assets/images/og-image-8b3e4f7d.png"
	}`, channelID)
}
