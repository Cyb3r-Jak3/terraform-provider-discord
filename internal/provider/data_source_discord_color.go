package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/go-playground/colors.v1"
	"strconv"
	"strings"
)

var _ datasource.DataSource = &DiscordColor{}

func NewDiscordColorDataSource() datasource.DataSource {
	return &DiscordColor{}
}

type DiscordColorModel struct {
	Hex types.String `tfsdk:"hex"`
	RGB types.String `tfsdk:"rgb"`
	Dec types.Int64  `tfsdk:"dec"`
}

type DiscordColor struct {
	client *Context
}

func (r *DiscordColor) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_color"

}

func (r *DiscordColor) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *DiscordColor) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Discord Color Data Source",
		Attributes: map[string]schema.Attribute{
			"hex": schema.StringAttribute{
				Description: "Hexadecimal color code",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("hex"), path.MatchRoot("rgb")),
				},
			},
			"rgb": schema.StringAttribute{
				Description: "RGB color code",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("hex"), path.MatchRoot("rgb")),
				},
			},
			"dec": schema.Int64Attribute{
				Description: "Decimal color code",
				Computed:    true,
			},
		},
	}
}

func (r *DiscordColor) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DiscordColorModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	var hex string
	if data.Hex.ValueString() != "" {
		if clr, err := colors.ParseHEX(data.Hex.ValueString()); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to parse hex %s", data.Hex), err.Error())
			return
		} else {
			hex = clr.String()
		}
	}
	if data.RGB.ValueString() != "" {
		if clr, err := colors.ParseRGB(data.RGB.ValueString()); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to parse rgb %s", data.RGB), err.Error())
			return
		} else {
			hex = clr.ToHEX().String()
		}
	}

	if intColor, err := ConvertToInt(hex); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to parse hex %s", hex), err.Error())
		return
	} else {
		data.Dec = types.Int64Value(intColor)

	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ConvertToInt(hex string) (int64, error) {
	hex = strings.Replace(hex, "0x", "", 1)
	hex = strings.Replace(hex, "0X", "", 1)
	hex = strings.Replace(hex, "#", "", 1)

	return strconv.ParseInt(hex, 16, 64)
}
