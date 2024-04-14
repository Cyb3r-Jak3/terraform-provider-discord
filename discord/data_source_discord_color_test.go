package discord

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDatasourceDiscordColor(t *testing.T) {
	name := "data.discord_color.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordColorRGB,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						name, "dec", "203569"),
				),
			},
			{
				Config: testAccDatasourceDiscordColorHex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						name, "dec", "203569"),
				),
			},
		},
	})
}

const testAccDatasourceDiscordColorHex = `
data "discord_color" "example" {
  hex = "#031b31"
}
`

const testAccDatasourceDiscordColorRGB = `
data "discord_color" "example" {
  rgb = "rgb(3, 27, 49)"
}
`
