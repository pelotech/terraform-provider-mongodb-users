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
