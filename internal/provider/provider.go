// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
  "os"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
  "github.com/hashicorp/terraform-plugin-framework/path"
  "github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &mongodbUsersProvider{}

type mongodbUsersProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type mongodbUsersProviderModel struct {
  Host types.String `tfsdk:"host"`
  Username types.String `tfsdk:"username"`
  Password types.String `tfsdk:"password"`
}


func (p *mongodbUsersProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mongodb-users"
	resp.Version = p.version
}

func (p *mongodbUsersProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
  resp.Schema = schema.Schema{
    Attributes: map[string]schema.Attribute{
        "host": schema.StringAttribute{
            Required: true,
        },
        "username": schema.StringAttribute{
            Required: true,
        },
        "password": schema.StringAttribute{
            Required: true,
            Sensitive: true,
        },
    },
  }
}

func (p *mongodbUsersProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Retrieve provider data from configuration
    var config mongodbUsersProviderModel
    diags := req.Config.Get(ctx, &config)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    if config.Host.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Unknown MongoDb API Host",
            "The provider cannot create the MongoDb API client as there is an unknown configuration value for the MongoDb API host. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the MONGODB_HOST environment variable.",
        )
    }

    if config.Username.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("username"),
            "Unknown MongoDb API Username",
            "The provider cannot create the MongoDb API client as there is an unknown configuration value for the MongoDb API username. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the MONGODB_USERNAME environment variable.",
        )
    }

    if config.Password.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("password"),
            "Unknown MongoDb API Password",
            "The provider cannot create the MongoDb API client as there is an unknown configuration value for the MongoDb API password. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the MONGODB_PASSWORD environment variable.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Default values to environment variables, but override
    // with Terraform configuration value if set.

    host := os.Getenv("MONGODB_HOST")
    username := os.Getenv("MONGODB_USERNAME")
    password := os.Getenv("MONGODB_PASSWORD")

    if !config.Host.IsNull() {
        host = config.Host.ValueString()
    }

    if !config.Username.IsNull() {
        username = config.Username.ValueString()
    }

    if !config.Password.IsNull() {
        password = config.Password.ValueString()
    }

    // If any of the expected configurations are missing, return
    // errors with provider-specific guidance.

    if host == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Missing MongoDb API Host",
            "The provider cannot create the MongoDb API client as there is a missing or empty value for the MongoDb API host. "+
                "Set the host value in the configuration or use the MONGODB_HOST environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if username == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("username"),
            "Missing MongoDb API Username",
            "The provider cannot create the MongoDb API client as there is a missing or empty value for the MongoDb API username. "+
                "Set the username value in the configuration or use the MONGODB_USERNAME environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if password == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("password"),
            "Missing MongoDb API Password",
            "The provider cannot create the MongoDb API client as there is a missing or empty value for the MongoDb API password. "+
                "Set the password value in the configuration or use the MONGODB_PASSWORD environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }


    mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    credential := options.Credential{
      AuthMechanism: "PLAIN",
      Username: username,
      Password: password,
    }

    client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb://" + host).
        SetAuth(credential))

    defer func() {
      if err = client.Disconnect(ctx); err != nil {
        panic(err)
      }
    }()




    // Make the MongoDb client available during DataSource and Resource
    // type Configure methods.
    resp.DataSourceData = client
    resp.ResourceData = client
}

func (p *mongodbUsersProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewUserResource,
    }
}

func (p *mongodbUsersProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
  return nil
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &mongodbUsersProvider{
            version: version,
        }
    }
}

