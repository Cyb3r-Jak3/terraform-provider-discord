package provider

import (
	"context"
	"fmt"
	"github.com/Cyb3r-Jak3/discord-terraform/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DiscordServerResource{}
var _ resource.ResourceWithImportState = &DiscordServerResource{}

func NewDiscordServerResource() resource.Resource {
	return &DiscordServerResource{}
}

type DiscordServerResource struct {
	client *Context
}

func (r *DiscordServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *DiscordServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{

		MarkdownDescription: "Discord Server Resource",
		Attributes:          utils.BuildServerResourceSchema(false),
	}
}

func (r *DiscordServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DiscordServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	utils.DiscordServerCreate(r.client.Session, ctx, req, resp)
}

func (r *DiscordServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	utils.DiscordServerRead(r.client.Session, ctx, req, resp)
}

func (r *DiscordServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	utils.DiscordServerUpdate(r.client.Session, ctx, req, resp)
}

func (r *DiscordServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	utils.DiscordServerDelete(r.client.Session, ctx, req, resp)
}

func (r *DiscordServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("server_id"), req, resp)
}
