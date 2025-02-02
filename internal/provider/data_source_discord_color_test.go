package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDatasourceDiscordColor(t *testing.T) {
	name := "data.discord_color.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordColorRGB,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						name, "dec", "203569"),
				),
			},
			{
				Config: testAccDatasourceDiscordColorHex,
				Check: resource.ComposeAggregateTestCheckFunc(
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
