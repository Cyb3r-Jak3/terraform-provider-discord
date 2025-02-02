package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordRole(t *testing.T) {
	testServerID := os.Getenv("DISCORD_TEST_SERVER_ID")
	if testServerID == "" {
		t.Skip("DISCORD_TEST_SERVER_ID envvar must be set for acceptance tests")
	}
	name := "discord_role.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordRole(testServerID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "server_id", testServerID),
					resource.TestCheckResourceAttr(name, "name", "terraform-test-role"),
					resource.TestCheckResourceAttr(name, "color", "65280"),
					resource.TestCheckResourceAttr(name, "hoist", "true"),
					resource.TestCheckResourceAttr(name, "mentionable", "true"),
					resource.TestCheckResourceAttr(name, "position", "1"),
					resource.TestCheckResourceAttr(name, "permissions", "1024"),
					resource.TestCheckResourceAttrSet(name, "id"),
				),
			},
			{
				ResourceName:        name,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s/", testServerID),
			},
		},
	})
}

func testAccResourceDiscordRole(channelID string) string {
	return fmt.Sprintf(`
    data "discord_color" "green" {
    	hex = "#00ff00"
	}

	resource "discord_role" "example" {
		server_id = "%[1]s"
        name = "terraform-test-role"
        color = data.discord_color.green.dec
        hoist = true
  	    mentionable = true
        position = 1
        permissions = 1024
	}`, channelID)
}
