terraform {
  required_providers {
    mongodb-users = {
      source = "hashicorp.com/edu/mongodb-users"
    }
  }
}

provider "mongodb-users" {

}

data "mongodb-user" "example" {}
