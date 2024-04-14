package discord

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceDiscordServer(t *testing.T) {
	name := "discord_server.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordServer,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "example"),
					resource.TestCheckResourceAttrSet(name, "region"),
					resource.TestCheckResourceAttr(name, "default_message_notifications", "0"),
					resource.TestCheckResourceAttr(name, "verification_level", "0"),
					resource.TestCheckResourceAttr(name, "explicit_content_filter", "0"),
					resource.TestCheckResourceAttr(name, "afk_timeout", "300"),
					resource.TestCheckResourceAttrSet(name, "owner_id"),
				),
			},
		},
	})
}

const testAccResourceDiscordServer = `
resource "discord_server" "example" {
  name = "example"
}
`
