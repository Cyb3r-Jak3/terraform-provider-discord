package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordInvite(t *testing.T) {
	testChannelID := os.Getenv("DISCORD_TEST_CHANNEL_ID")
	if testChannelID == "" {
		t.Skip("DISCORD_TEST_CHANNEL_ID envvar must be set for acceptance tests")
	}
	name := "discord_invite.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordInvite(testChannelID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "channel_id", testChannelID),
					resource.TestCheckResourceAttr(name, "max_age", "86400"),
					resource.TestCheckResourceAttr(name, "max_uses", "1"),
					resource.TestCheckResourceAttr(name, "temporary", "true"),
					resource.TestCheckResourceAttr(name, "unique", "false"),
					resource.TestCheckResourceAttrSet(name, "code"),
				),
			},
			{
				ResourceName:        name,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s/", testChannelID),
			},
		},
	})
}

func testAccResourceDiscordInvite(channelID string) string {
	return fmt.Sprintf(`
	resource "discord_invite" "example" {
      channel_id = "%[1]s"
      max_age = 86400
      max_uses = 1
      temporary = true
      unique = false
	}`, channelID)
}
