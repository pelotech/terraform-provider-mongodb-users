resource "mongodb-users_user" "user1" {
  user     = "user1"
  db       = "test"
  password = "abc123"
  roles = [
    {
      db   = "test"
      role = "readWrite"
    }
  ]
}
