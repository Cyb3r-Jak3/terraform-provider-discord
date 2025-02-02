package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccResourceDiscordMessageContent(t *testing.T) {
	testChannelID := os.Getenv("DISCORD_TEST_CHANNEL_ID")
	if testChannelID == "" {
		t.Skip("DISCORD_TEST_CHANNEL_ID envvar must be set for acceptance tests")
	}
	name := "discord_message.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordMessageContent(testChannelID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "channel_id", testChannelID),
					resource.TestCheckResourceAttr(name, "content", "Hello, World from Terraform!"),
					resource.TestCheckResourceAttr(name, "tts", "false"),
				),
			},
			{
				ResourceName:        name,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s/", testChannelID),
			},
		},
	})
}

func TestAccResourceDiscordMessageEmbed(t *testing.T) {
	testChannelID := os.Getenv("DISCORD_TEST_CHANNEL_ID")
	if testChannelID == "" {
		t.Skip("DISCORD_TEST_CHANNEL_ID envvar must be set for acceptance tests")
	}
	name := "discord_message.example"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDiscordEmbed(testChannelID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "channel_id", testChannelID),
					resource.TestCheckResourceAttr(name, "embed.#", "1"),
					resource.TestCheckResourceAttr(name, "embed.0.title", "Hello, World from Terraform! (Embed)"),
					resource.TestCheckResourceAttr(name, "embed.0.description", "This is a test emebed from Terraform!"),
					resource.TestCheckResourceAttr(name, "embed.0.color", "65280"),
					resource.TestCheckResourceAttr(name, "embed.0.footer.text", "This is a test footer from Terraform!"),
				),
			},
			{
				ResourceName:        name,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s/", testChannelID),
			},
		},
	})
}

func testAccResourceDiscordMessageContent(channelID string) string {
	return fmt.Sprintf(`
	resource "discord_message" "example" {
      channel_id = "%[1]s"
      content = "Hello, World from Terraform!"
	  tts = false
	}`, channelID)
}

func testAccResourceDiscordEmbed(channelID string) string {
	return fmt.Sprintf(`
    data "discord_color" "green" {
    	hex = "#00ff00"
		}
	resource "discord_message" "example" {
      channel_id = "%[1]s"
      embed {
           title = "Hello, World from Terraform! (Embed)"
           description = "This is a test emebed from Terraform!"
 		   color = data.discord_color.green.dec
 		   footer {
              text = "This is a test footer from Terraform!"
		   }
		}
	}`, channelID)
}
