package discord

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestAccDatasourceDiscordServer(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID is not set")
	}

	name := "data.discord_server.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordServer(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Discord Terraform Test Server"),
					resource.TestCheckResourceAttrSet(name, "region"),
					resource.TestCheckResourceAttr(name, "default_message_notifications", "1"),
					resource.TestCheckResourceAttr(name, "verification_level", "1"),
					resource.TestCheckResourceAttr(name, "explicit_content_filter", "2"),
					resource.TestCheckResourceAttr(name, "afk_timeout", "300"),
					resource.TestCheckResourceAttrSet(name, "owner_id"),
				),
			},
		},
	})
}

func testAccDatasourceDiscordServer(serverId string) string {
	return fmt.Sprintf(`
	data "discord_server" "example" {
	  server_id = "%[1]s"
	}`, serverId)
}
