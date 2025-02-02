package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"
)

func TestAccResourceDiscordServer(t *testing.T) {
	name := "discord_server.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
			{
				ResourceName:                         name,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    BuildImportStateIdFunc(name, "server_id"),
				ImportStateVerifyIdentifierAttribute: "server_id",
			},
		},
	})
}

const testAccResourceDiscordServer = `
resource "discord_server" "example" {
  name = "example"
}
`

// BuildImportStateIdFunc constructs a function that returns the id attribute of a target resouce from the terraform state.
// This is a helper function for conveniently constructing the ImportStateIdFunc field for a test step.
func BuildImportStateIdFunc(resourceId, attr string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		// Find the resource in the Terraform state.
		rs, ok := s.RootModule().Resources[resourceId]
		if !ok {
			return "", fmt.Errorf("resource not found in state: %s", resourceId)
		}

		// Access the attribute directly from the state.
		targetAttr := rs.Primary.Attributes[attr]
		if targetAttr == "" {
			return "", fmt.Errorf("attribute '%s' not found or empty in the resource", attr)
		}

		// Return the found attribute or the ID needed for the import.
		return targetAttr, nil
	}
}
