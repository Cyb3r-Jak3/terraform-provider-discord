---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "discord_role Data Source - discord"
subcategory: ""
description: |-
  Discord Role
---

# discord_role (Data Source)

Discord Role

## Example Usage

```terraform
data "discord_role" "mods_id" {
  server_id = "81384788765712384"
  role_id   = "175643578071121920"
}

data "discord_role" "mods_name" {
  server_id = "81384788765712384"
  name      = "Mods"
}

output "mods_color" {
  value = data.discord_role.mods_id.color
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `server_id` (String)

### Optional

- `name` (String)

### Read-Only

- `color` (Number)
- `hoist` (Boolean)
- `id` (String) The ID of this resource.
- `managed` (Boolean)
- `mentionable` (Boolean)
- `permissions` (Number)
- `position` (Number)


