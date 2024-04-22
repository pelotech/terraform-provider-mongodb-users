package provider

import (
    "context"
    "time"
    "fmt"
    "errors"
    "strings"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

var (
    _ resource.Resource = &userResource{}
    _ resource.ResourceWithConfigure = &userResource{}
    _ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
  return &userResource{}
}

type userResource struct {
  client *mongo.Client
}

type userResourceModel struct {
  Id types.String `tfsdk:"id"`
  User types.String `tfsdk:"user"`
  Password types.String `tfsdk:"password"`
  Db types.String `tfsdk:"db"`
  Roles []userRoleModel `tfsdk:"roles"`
  LastUpdated types.String `tfsdk:"last_updated"`
}

type userRoleModel struct {
  Db types.String `tfsdk:"db"`
  Role types.String `tfsdk:"role"`
}

type commandResponse struct {
  OK            int       `bson:"ok"`
  OperationTime time.Time `bson:"operationTime"`
}

type dbUser struct {
  Id         string `bson:"_id"`
  User       string `bson:"user"`
  Db         string `bson:"db"`
  Roles      []dbRole `bson:"roles"`
}

type dbRole struct {
	Role string `bson:"role"`
	Db   string `bson:"db"`
}

type readResponse struct {
  commandResponse `bson:",inline"`
  Users []dbUser `bson:"users"`
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
      Attributes: map[string]schema.Attribute{
        "id": schema.StringAttribute{
          Computed: true,
        },
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
        "roles": schema.SetAttribute{
          ElementType: types.ObjectType{
            AttrTypes: map[string]attr.Type{
              "db" : types.StringType,
              "role" : types.StringType,
            },
          },
          Required: true,
        },
        "last_updated" :schema.StringAttribute{
          Computed: true,
        },
      },
    }
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan userResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
      return
    }

    var roles []bson.M
    for _, role := range plan.Roles {
      roles = append(roles, bson.M {"role": role.Role.ValueString(), "db": role.Db.ValueString()})
    }

    userCreateCommand := bson.D {{"createUser", plan.User.ValueString()}, {"pwd", plan.Password.ValueString()}, {"roles", roles}}

    mongoResult := r.client.Database(plan.Db.ValueString()).RunCommand(ctx, userCreateCommand)
    if mongoResult.Err() != nil {
      resp.Diagnostics.AddError(
          "Error creating user",
          "Could not create user, unexpected error: " + mongoResult.Err().Error(),
      )
      return
    }

    var response commandResponse
    err := mongoResult.Decode(&response)
    if err != nil {
      resp.Diagnostics.AddError(
        "Error creating user",
        "Could not create user, unexpected error: " + err.Error(),
      )

      return
    }

    if response.OK != 1 {
      resp.Diagnostics.AddError(
        "Error creating user",
        fmt.Sprintf("Could not create user, unexpected error returned from MongoDB: %d", response.OK))
      return
    }

    // Read back user from DB to get ID
    user, err := r.getUserFromDb(ctx, plan.Db.ValueString(), plan.User.ValueString())
    if err != nil {
      resp.Diagnostics.AddError(
          "Error reading user from MongoDb",
          "Could not retrieve user <" + plan.User.ValueString() + "> " + err.Error())
    }

    // Set state to fully populated data
    plan.Id = types.StringValue(user.Id)
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *userResource) getUserFromDb(ctx context.Context, db string, user string) (dbUser, error) {
  var usersInfo readResponse
  cmd := bson.D{{Key: "usersInfo", Value: bson.M{
      "user": user,
      "db":   db,
  }}}

  err := r.client.Database(db).RunCommand(ctx, cmd).Decode(&usersInfo)
  if err != nil {
    return dbUser{}, err
  }

  users := usersInfo.Users
  if len(users) == 0 {
    return dbUser{}, errors.New("No users matched for db: " + db + " and user: " + user)
  }

  return users[0], nil
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
  var state userResourceModel
  diags := req.State.Get(ctx, &state)
  resp.Diagnostics.Append(diags...)
  if resp.Diagnostics.HasError() {
    return
  }

  user, err := r.getUserFromDb(ctx, state.Db.ValueString(), state.User.ValueString())
  if err != nil {
    resp.Diagnostics.AddError(
        "Error reading user from MongoDb",
        "Could not retrieve user <" + state.User.ValueString() + "> " + err.Error())
  }

  state.Id = types.StringValue(user.Id)
  state.User = types.StringValue(user.User)
  state.Db = types.StringValue(user.Db)

  state.Roles = []userRoleModel{}
  for _, item := range user.Roles {
    state.Roles = append(state.Roles, userRoleModel {
      Db: types.StringValue(item.Db),
      Role: types.StringValue(item.Role),
    })
  }

  diags = resp.State.Set(ctx, &state)
  resp.Diagnostics.Append(diags...)
  if (resp.Diagnostics.HasError()) {
    return
  }
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan userResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
      return
    }

    var roles []bson.M
    for _, role := range plan.Roles {
      roles = append(roles, bson.M {"role": role.Role.ValueString(), "db": role.Db.ValueString()})
    }
    userUpdateCommand := bson.D {{"updateUser", plan.User.ValueString()}, {"pwd", plan.Password.ValueString()}, {"roles", roles}}

    mongoResult := r.client.Database(plan.Db.ValueString()).RunCommand(ctx, userUpdateCommand)
    if mongoResult.Err() != nil {
      resp.Diagnostics.AddError(
          "Error updating user",
          "Could not update user, unexpected error: " + mongoResult.Err().Error(),
      )
      return
    }

    var response commandResponse
    err := mongoResult.Decode(&response)
    if err != nil {
      resp.Diagnostics.AddError(
        "Error updating user",
        "Could not update user, unexpected error: " + err.Error(),
      )

      return
    }

    if response.OK != 1 {
      resp.Diagnostics.AddError(
        "Error updating user",
        fmt.Sprintf("Could not update user, unexpected error returned from MongoDB: %d", response.OK))
      return
    }

    // Read back user from DB to get ID
    user, err := r.getUserFromDb(ctx, plan.Db.ValueString(), plan.User.ValueString())
    if err != nil {
      resp.Diagnostics.AddError(
          "Error reading user from MongoDb",
          "Could not retrieve user <" + plan.User.ValueString() + "> " + err.Error())
    }

    // Set state to fully populated data
    plan.Id = types.StringValue(user.Id)
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
  var state userResourceModel
  diags := req.State.Get(ctx, &state)
  resp.Diagnostics.Append(diags...)
  if resp.Diagnostics.HasError() {
    return
  }

  userDeleteCommand := bson.D {{"dropUser", state.User.ValueString()}}

  mongoResult := r.client.Database(state.Db.ValueString()).RunCommand(ctx, userDeleteCommand)
  if mongoResult.Err() != nil {
    resp.Diagnostics.AddError(
        "Error deleting user",
        "Could not delete user, unexpected error: " + mongoResult.Err().Error(),
    )
    return
  }
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    idParts := strings.SplitN(req.ID, ".", 2)

    if len (idParts) < 2 {
      resp.Diagnostics.AddError(
          "Unexpected import identifier",
          fmt.Sprintf("Expected import identifier with format: <db>.<user>  Got: %q", req.ID),
      )
      return
    }

    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("db"), idParts[0])...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user"), idParts[1])...)
}
