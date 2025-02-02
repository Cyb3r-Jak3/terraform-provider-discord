package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDatasourceDiscordLocalImage(t *testing.T) {
	name := "data.discord_local_image.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordLocalImage,
				Check: resource.ComposeAggregateTestCheckFunc(
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
