---
page_title: "discord Provider"
subcategory: ""
description: |-
  
---

# discord Provider

This is a fork of [Lucky3028/terraform-provider-discord](https://github.com/Lucky3028/terraform-provider-discord).

The Discord provider is used to interact with the Discord API. It requires proper credentials before it can be used.

Use the navigation on the left to read more about the resources and data sources.

## Example Usage

```terraform
provider "discord" {
  token = var.discord_token
}

data "discord_local_image" "logo" {
  file = "logo.png"
}

resource "discord_server" "my_server" {
  name                          = "My Awesome Server"
  region                        = "us-west"
  default_message_notifications = 0
  icon_data_uri                 = data.discord_local_image.logo.data_uri
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `client_id` (String)
- `secret` (String)
- `token` (String) Discord API Token. This can be found in the Discord Developer Portal. This includes the `Bot` prefix. Can also be set via the `DISCORD_TOKEN` environment variable.
