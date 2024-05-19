package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

// Ensure DiscordProvider satisfies various provider interfaces.
var _ provider.Provider = &DiscordProvider{}
var _ provider.ProviderWithFunctions = &DiscordProvider{}

// DiscordProvider defines the provider implementation.
type DiscordProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// DiscordProviderModel describes the provider data model.
type DiscordProviderModel struct {
	Token    types.String `tfsdk:"token"`
	ClientID types.String `tfsdk:"client_id"`
	Secret   types.String `tfsdk:"secret"`
}

func (p *DiscordProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "discord"
	resp.Version = p.version
}

func (p *DiscordProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "Discord API Token. This can be found in the Discord Developer Portal. This includes the `Bot` prefix. Can also be set via the `DISCORD_TOKEN` environment variable.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				Optional: true,
			},
			"secret": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *DiscordProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DiscordProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	token := data.Token.ValueString()
	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
	}
	if token == "" {
		resp.Diagnostics.AddError("missing required token", "the `token` argument or `DISCORD_TOKEN` environment variable must be set")
		return
	}
	config := Config{
		Token:    token,
		ClientID: data.ClientID.ValueString(),
		Secret:   data.Secret.ValueString(),
	}

	client, err := config.Client(p.version)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Discord client", err.Error())
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DiscordProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDiscordInviteResource,
		NewDiscordMessageResource,
		NewDiscordRoleResource,
		NewDiscordServerResource,
		NewDiscordManagedServerResource,
		NewDiscordVoiceChannelResource,
		NewDiscordNewsChannelResource,
		NewDiscordCategoryChannelResource,
		NewDiscordTextChannelResource,
		//NewDiscordEveryoneRoleResource,
		NewDiscordSystemChannelResource,
		NewDiscordWebhookResource,
		NewDiscordChannelPermissionResource,
	}
}

func (p *DiscordProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDiscordRoleDataSource,
		NewDiscordColorDataSource,
		NewDiscordLocalImageDataSource,
		NewDiscordMemberDataSource,
		NewDiscordPermissionDataSource,
		NewDiscordServerDataSource,
		NewDiscordSystemChannelDataSource,
	}
}

func (p *DiscordProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DiscordProvider{
			version: version,
		}
	}
}
