package provider

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
    _ resource.Resource = &userResource{}
    _ resource.ResourceWithConfigure = &userResource{}
)

func NewUserResource() resource.Resource {
  return &userResource{}
}

type userResource struct {
  client *mongo.Client
}

type userResourceModel struct {
  User types.String `tfsdk:"user"`
  Password types.String `tfsdk:"password"`
  Db types.String `tfsdk:"db"`
  Roles []userRoleModel `tfsdk:"roles"`
}

type userRoleModel struct {
  Db types.String `tfsdk:"db"`
  Role types.String `tfsdk:"role"`
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*mongo.Client)

    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *mongo.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )

        return
    }

    r.client = client
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
      Attributes: map[string].schema.Attribute{
        "db": schema.StringAttribute{
          Required: true,
        },
        "user": schema.StringAttribute{
          Required: true,
        },
        "password": schema.StringAttribute{
          Required: true,
          Sensitive: true,
        },
        "roles": schema.ListNestedAttribute{
          Required: true,
          NestedObject: schema.NestedAttributeObject {
            Attributes: map[string]schema.Attribute{
              "db": schema.StringAttribute{
                Required: true,
              },
              "role": schema.StringAttribute{
                Required: true,
              },
            },
          },
        },
      }
    }
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan userResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diag...)
    if resp.Diagnostics.HasError() {
      return
    }

    var roles []bson.M
    for _, role := range plan.Roles {
      roles = append(roles, bson.M{{"role": role.Role, "db": role.Db}})
    }

    userCreateCommand := bson.D {{"createUser", plan.Name}, {"pwd", plan.Password}, {"roles", roles}}

    mongoResult := client.Database(plan.Db).RunCommand(context, userCreateCommand)
    if mongoResult.Err() != nil {
      resp.Diagnostics.AddError(
          "Error creating user",
          "Could not create user, unexpected error: " + mongoReuslt.Err(),
      )
      return
    }

    plan.ID = types.StringValue(str

    // Map response body to schema and populate Computed attribute values
    plan.ID = types.StringValue(strconv.Itoa(order.ID))
    for orderItemIndex, orderItem := range order.Items {
        plan.Items[orderItemIndex] = orderItemModel{
            Coffee: orderItemCoffeeModel{
                ID:          types.Int64Value(int64(orderItem.Coffee.ID)),
                Name:        types.StringValue(orderItem.Coffee.Name),
                Teaser:      types.StringValue(orderItem.Coffee.Teaser),
                Description: types.StringValue(orderItem.Coffee.Description),
                Price:       types.Float64Value(orderItem.Coffee.Price),
                Image:       types.StringValue(orderItem.Coffee.Image),
            },
            Quantity: types.Int64Value(int64(orderItem.Quantity)),
        }
    }
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    // Set state to fully populated data
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}




func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
