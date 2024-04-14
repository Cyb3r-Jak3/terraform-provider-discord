package discord

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDatasourceDiscordLocalImage(t *testing.T) {
	name := "data.discord_local_image.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordLocalImage,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "file", "provider.go"),
					resource.TestCheckResourceAttrSet(name, "data_uri"),
				),
			},
		},
	})
}

const testAccDatasourceDiscordLocalImage = `
data "discord_local_image" "example" {
  file = "provider.go"
}
`
