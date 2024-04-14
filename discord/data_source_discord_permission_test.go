package discord

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDatasourceDiscordPermission(t *testing.T) {
	name := "data.discord_permission.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordPermissionSimple,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "administrator", "allow"),
					resource.TestCheckResourceAttr(name, "allow_bits", "8"),
				),
			},
			{
				Config: testAccDatasourceDiscordPermissionComplex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "send_messages", "allow"),
					resource.TestCheckResourceAttr(name, "embed_links", "allow"),
					resource.TestCheckResourceAttr(name, "allow_bits", "18432"),
					resource.TestCheckResourceAttr(name, "speak", "deny"),
					resource.TestCheckResourceAttr(name, "change_nickname", "deny"),
					resource.TestCheckResourceAttr(name, "deny_bits", "69206016"),
				),
			},
		},
	})
}

const testAccDatasourceDiscordPermissionSimple = `
data "discord_permission" "example" {
  administrator = "allow"
}
`

const testAccDatasourceDiscordPermissionComplex = `
data "discord_permission" "example" {
  send_messages = "allow"
	  embed_links = "allow"
  speak = "deny"
	change_nickname = "deny"
}
`
