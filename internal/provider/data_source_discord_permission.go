package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"reflect"
	"strings"
)

var Permissions = map[string]int64{
	"create_instant_invite":       0x1,
	"kick_members":                0x2,
	"ban_members":                 0x4,
	"administrator":               0x8,
	"manage_channels":             0x10,
	"manage_guild":                0x20,
	"add_reactions":               0x40,
	"view_audit_log":              0x80,
	"priority_speaker":            0x100,
	"stream":                      0x200,
	"view_channel":                0x400,
	"send_messages":               0x800,
	"send_tts_messages":           0x1000,
	"manage_messages":             0x2000,
	"embed_links":                 0x4000,
	"attach_files":                0x8000,
	"read_message_history":        0x10000,
	"mention_everyone":            0x20000,
	"use_external_emojis":         0x40000,
	"view_guild_insights":         0x80000,
	"connect":                     0x100000,
	"speak":                       0x200000,
	"mute_members":                0x400000,
	"deafen_members":              0x800000,
	"move_members":                0x1000000,
	"use_vad":                     0x2000000,
	"change_nickname":             0x4000000,
	"manage_nicknames":            0x8000000,
	"manage_roles":                0x10000000,
	"manage_webhooks":             0x20000000,
	"manage_emojis":               0x40000000,
	"use_application_commands":    0x80000000,
	"request_to_speak":            0x100000000,
	"manage_events":               0x200000000,
	"manage_threads":              0x400000000,
	"create_public_threads":       0x800000000,
	"create_private_threads":      0x1000000000,
	"use_external_stickers":       0x2000000000,
	"send_thread_messages":        0x4000000000,
	"start_embedded_activities":   0x8000000000,
	"moderate_members":            0x10000000000,
	"view_monetization_analytics": 0x20000000000,
	"use_soundboard":              0x40000000000,
	"create_expressions":          0x80000000000,
	"create_events":               0x100000000000,
	"use_external_sounds":         0x200000000000,
	"send_voice_messages":         0x400000000000,
	"set_voice_stats":             0x800000000000,
	"use_external_apps":           0x0004000000000000,
	"set_voice_channel_status":    0x0001000000000000,
}

var _ datasource.DataSource = &DiscordPermission{}

func NewDiscordPermissionDataSource() datasource.DataSource {
	return &DiscordPermission{}
}

type DiscordPermission struct {
	client *Context
}

func (r *DiscordPermission) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"

}

func (r *DiscordPermission) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Context)

	if !ok {
		resp.Diagnostics.AddError(
			"unexpected resource configure type",
			fmt.Sprintf("Expected *Context, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *DiscordPermission) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Discord Color Data Source",
		Attributes: map[string]schema.Attribute{
			"allow_extends": schema.Int64Attribute{
				Optional: true,
			},
			"deny_extends": schema.Int64Attribute{
				Optional: true,
			},
			"allow_bits": schema.Int64Attribute{
				Computed: true,
			},
			"deny_bits": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
	for k := range Permissions {
		resp.Schema.Attributes[k] = schema.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf("allow", "unset", "deny"),
			},
		}

	}
}

func (r *DiscordPermission) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DiscordPermissionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	var allowBits int64
	var denyBits int64
	v := reflect.ValueOf(data)
	for perm, bit := range Permissions {
		correctedName := strings.ReplaceAll(strings.Title(strings.ReplaceAll(perm, "_", " ")), " ", "")

		switch (v.FieldByName(correctedName).Interface().(basetypes.StringValue)).ValueString() {
		case "allow":
			allowBits |= bit
		case "deny":
			denyBits |= bit
		}
	}
	data.AllowBits = types.Int64Value(allowBits | data.AllowExtends.ValueInt64())
	data.DenyBits = types.Int64Value(denyBits | data.DenyExtends.ValueInt64())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
