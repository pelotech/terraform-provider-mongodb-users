terraform {
  required_providers {
    mongodb-users = {
      source = "hashicorp.com/edu/mongodb-users"
    }
  }
}

provider "mongodb-users" {
  host     = "localhost:27017"
  username = "root"
  password = "password123"
}

resource "mongodb-users_user" "junky" {
  user = "unky"
  db = "test"
  password = "123password"
  roles = [
    {
      db = "test"
      role = "readWrite"
    }
  ]
}
