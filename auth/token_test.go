package auth

import (
  "testing"
  "time"

  "github.com/jimmc/mimsrv/users"
)

func TestIsValid(t *testing.T) {
  initTokens()
  if isValidToken("user1", "id1") {
    t.Fatal("token was deemed valid before any tokens added")
  }
  user1 := users.NewUser("user1", "cw1", nil)
  token := newToken(user1, "id1")
  if !isValidToken(token.Key, "id1") {
    t.Fatalf("Token %s should be valid", token.Key)
  }
  if isValidToken(token.Key, "id2") {
    t.Fatalf("Token %s with different idstr should be invalid", token.Key)
  }
  if isValidToken("user2", "id2") {
    t.Fatal("token was deemed valid before being created")
  }

  timeNow = func() time.Time { return time.Now().Add(time.Hour * 30) }
  if isValidToken(token.Key, "id1") {
    t.Fatalf("Token %s should be invalid after timeout", token.Key)
  }
}
