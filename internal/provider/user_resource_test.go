package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "mongodb-users_user" "test_1" {
  user = "test_1"
  db = "test"
  password = "test1"
  roles = [
    {
      db = "test"
      role = "readWrite"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "user", "test_1"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "db", "test"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.#", "1"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.0.db", "test"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.0.role", "readWrite"),

					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("mongodb-users_user.test_1", "id"),
					resource.TestCheckResourceAttrSet("mongodb-users_user.test_1", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName: "mongodb-users_user.test_1",

				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "test.test_1",
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "password"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "mongodb-users_user" "test_1" {
  user = "test_1"
  db = "test"
  password = "test1"
  roles = [
    {
      db = "test"
      role = "readWrite"
    },
    {
      db = "test_other"
      role = "read"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "user", "test_1"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "db", "test"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.#", "2"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.0.db", "test"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.0.role", "readWrite"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.1.db", "test_other"),
					resource.TestCheckResourceAttr("mongodb-users_user.test_1", "roles.1.role", "read"),

					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("mongodb-users_user.test_1", "id"),
					resource.TestCheckResourceAttrSet("mongodb-users_user.test_1", "last_updated"),
				),
			},
		},
	})
}
