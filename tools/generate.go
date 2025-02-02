package main

import (
	"fmt"
	"github.com/Cyb3r-Jak3/discord-terraform/internal/provider"
	"os"
	"strings"
	"text/template"
	"time"
)

//go:generate go run generate.go
//go:generate gofmt -w ../internal/provider/

type TemplateData struct {
	Timestamp      string
	PermissionData string
}

type ChannelTemplateData struct {
	Timestamp           string
	ChannelType         string
	ModelName           string
	MarkdownDescription string
	ResourceName        string
	CanHaveParent       bool
	CanHaveTopic        bool
	CanHaveNSFW         bool
}

func main() {
	timestamp := time.Now().Format(time.RFC3339)
	generatePermission(timestamp)
	generateChannels(timestamp)
}

func generatePermission(timestamp string) {
	// Write the permissions map to a struct in permissions_model.go
	templateFile := "permission_model.go.tmpl"
	tmpl := template.Must(template.ParseFiles(templateFile))
	var permissionsStruct string
	for k, _ := range provider.Permissions {
		correctedName := strings.ReplaceAll(strings.Title(strings.ReplaceAll(k, "_", " ")), " ", "")
		permissionsStruct += fmt.Sprintf("%s types.String `tfsdk:\"%s\"`\n", correctedName, k)
	}
	f, err := os.Create("../internal/provider/data_source_discord_permissions_model.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = tmpl.Execute(f, TemplateData{
		Timestamp:      timestamp,
		PermissionData: permissionsStruct,
	})
	if err != nil {
		panic(err)
	}
}
func generateChannels(timestamp string) {
	templateFile := "channel_resource.go.tmpl"
	tmpl := template.Must(template.ParseFiles(templateFile))
	var channels = []ChannelTemplateData{
		{
			Timestamp:           timestamp,
			ChannelType:         "Text",
			ModelName:           "DiscordTextChannel",
			MarkdownDescription: "Discord Text Channel Resource",
			CanHaveParent:       true,
			CanHaveTopic:        true,
			CanHaveNSFW:         true,
		},
		{
			Timestamp:           timestamp,
			ChannelType:         "Voice",
			ModelName:           "DiscordVoiceChannel",
			MarkdownDescription: "Discord Voice Channel Resource",
			CanHaveParent:       true,
			CanHaveTopic:        false,
			CanHaveNSFW:         true,
		},
		{
			Timestamp:           timestamp,
			ChannelType:         "Category",
			ModelName:           "DiscordCategoryChannel",
			MarkdownDescription: "Discord Category Channel Resource",
			CanHaveParent:       false,
			CanHaveTopic:        false,
			CanHaveNSFW:         false,
		},
		{
			Timestamp:           timestamp,
			ChannelType:         "News",
			ModelName:           "DiscordNewsChannel",
			MarkdownDescription: "Discord News Channel Resource",
			CanHaveParent:       true,
			CanHaveTopic:        true,
			CanHaveNSFW:         false,
		},
		{
			Timestamp:           timestamp,
			ChannelType:         "Forum",
			ModelName:           "DiscordForumChannel",
			MarkdownDescription: "Discord Forum Channel Resource",
			CanHaveParent:       true,
			CanHaveTopic:        true,
			CanHaveNSFW:         false,
		},
	}
	for _, channel := range channels {
		channel.ResourceName = strings.ToLower(channel.ChannelType)
		f, err := os.Create(fmt.Sprintf("../internal/provider/resource_discord_channel_%s.go", channel.ResourceName))
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(f, channel)
		f.Close()
		if err != nil {
			panic(err)
		}

	}
}
