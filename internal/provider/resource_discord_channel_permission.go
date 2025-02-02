package provider

import (
	"context"
	"fmt"
	"github.com/Cyb3r-Jak3/discord-terraform/internal/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordChannelPermissionResource{}

//var _ resource.ResourceWithImportState = &DiscordChannelPermissionResource{}

func NewDiscordChannelPermissionResource() resource.Resource {
	return &DiscordChannelPermissionResource{}
}

type DiscordChannelPermissionResource struct {
	client *Context
}

type DiscordChannelPermissionModel struct {
	ChannelID   types.String `tfsdk:"channel_id"`
	Type        types.String `tfsdk:"type"`
	OverwriteID types.String `tfsdk:"overwrite_id"`
	Allow       types.Int64  `tfsdk:"allow"`
	Deny        types.Int64  `tfsdk:"deny"`
}

func (r *DiscordChannelPermissionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel_permission"
}

func (r *DiscordChannelPermissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Discord channel permission overwrite Resource",

		Attributes: map[string]schema.Attribute{
			"channel_id": schema.StringAttribute{
				Description: "The channel ID",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of permission overwrite",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("role", "member"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"overwrite_id": schema.StringAttribute{
				Description: "The ID of the role or member to overwrite",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allow": schema.Int64Attribute{
				Description: "The permissions to allow",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeastOneOf(path.MatchRoot("allow"), path.MatchRoot("deny")),
				},
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"deny": schema.Int64Attribute{
				Description: "The permissions to deny",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeastOneOf(path.MatchRoot("allow"), path.MatchRoot("deny")),
				},
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
		},
	}
}

func (r *DiscordChannelPermissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordChannelPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DiscordChannelPermissionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session

	channelID := data.ChannelID.ValueString()
	overrideID := data.OverwriteID.ValueString()
	permissionType, _ := utils.GetDiscordChannelPermissionType(data.Type.ValueString())
	if err := client.ChannelPermissionSet(
		channelID, overrideID, permissionType,
		data.Allow.ValueInt64(),
		data.Deny.ValueInt64(), discordgo.WithContext(ctx),
	); err != nil {
		resp.Diagnostics.AddError("Failed to set channel permission", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordChannelPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DiscordChannelPermissionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	channel, err := client.Channel(data.ChannelID.ValueString(), discordgo.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error fetching channel %s", data.ChannelID.ValueString()), err.Error())
		return

	}
	overrideID := data.OverwriteID.ValueString()
	permissionType, _ := utils.GetDiscordChannelPermissionType(data.Type.ValueString())

	found := false
	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.ID == overrideID && overwrite.Type == permissionType {
			permissionTypeOverwrite, _ := utils.GetDiscordChannelPermissionTypeString(overwrite.Type)
			data = &DiscordChannelPermissionModel{
				ChannelID:   types.StringValue(channel.ID),
				Type:        types.StringValue(permissionTypeOverwrite),
				OverwriteID: types.StringValue(overwrite.ID),
				Allow:       types.Int64Value(overwrite.Allow),
				Deny:        types.Int64Value(overwrite.Deny),
			}
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError("Failed to find permission overwrite", "Not found")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DiscordChannelPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DiscordChannelPermissionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session

	channelID := data.ChannelID.ValueString()
	overrideID := data.OverwriteID.ValueString()
	permissionType, _ := utils.GetDiscordChannelPermissionType(data.Type.ValueString())
	if err := client.ChannelPermissionSet(
		channelID, overrideID, permissionType,
		data.Allow.ValueInt64(),
		data.Deny.ValueInt64(), discordgo.WithContext(ctx),
	); err != nil {
		resp.Diagnostics.AddError("Failed to update channel permission overwrite", err.Error())
		return
	}

}

func (r *DiscordChannelPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DiscordChannelPermissionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	client := r.client.Session
	if err := client.ChannelPermissionDelete(data.ChannelID.ValueString(), data.OverwriteID.ValueString(), discordgo.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete channel permissions. channel: %s", data.ChannelID.ValueString()), err.Error())
		return
	}
}

//func (r *DiscordChannelPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
//	idparts := strings.Split(req.ID, "/")
//	fmt.Printf("ID parts: %v\n", idparts)
//	if len(idparts) != 3 {
//		resp.Diagnostics.AddError("Invalid ID", "Needed format: <channel_id>/<type>/id")
//		return
//	}
//	resp.Diagnostics.Append(resp.State.SetAttribute(
//		ctx, path.Root("channel_id"), idparts[0],
//	)...)
//	resp.Diagnostics.Append(resp.State.SetAttribute(
//		ctx, path.Root("type"), idparts[1],
//	)...)
//	resp.Diagnostics.Append(resp.State.SetAttribute(
//		ctx, path.Root("id"), idparts[2],
//	)...)
//}
