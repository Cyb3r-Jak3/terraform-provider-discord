package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccDatasourceDiscordSystemChannel(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}

	name := "data.discord_system_channel.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDiscordSystemChannel(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "server_id", testServerID),
					resource.TestCheckResourceAttrSet(name, "system_channel_id"),
				),
			},
		},
	})
}

func testAccDatasourceDiscordSystemChannel(serverId string) string {
	return fmt.Sprintf(`
	data "discord_system_channel" "example" {
	  server_id = "%[1]s"
	}`, serverId)
}
