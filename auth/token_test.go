package auth

import (
  "testing"
  "time"

  "github.com/jimmc/mimsrv/users"
)

func TestIsValid(t *testing.T) {
  initTokens()
  if _, v := currentToken("user1", "id1"); v {
    t.Fatal("token was deemed valid before any tokens added")
  }
  user1 := users.NewUser("user1", "cw1", nil)
  token := newToken(user1, "id1")
  var tk *Token
  var v bool
  if tk, v = currentToken(token.Key, "id1"); !v {
    t.Fatalf("Token %s should be valid", token.Key)
  }
  if tk != token {
    t.Fatalf("Token %s should be unique", token.Key)
  }
  if _, v := currentToken(token.Key, "id2"); v {
    t.Fatalf("Token %s with different idstr should be invalid", token.Key)
  }
  if _, v := currentToken("user2", "id2"); v {
    t.Fatal("token was deemed valid before being created")
  }

  timeNow = func() time.Time { return time.Now().Add(time.Hour * 30) }
  if _, v := currentToken(token.Key, "id1"); v {
    t.Fatalf("Token %s should be invalid after timeout", token.Key)
  }
}

func TestRefresh(t *testing.T) {
  initTokens()
  user2 := users.NewUser("user2", "cw2", nil)

  timeNow = func() time.Time { return time.Now() }
  token := newToken(user2, "id2")
  var v bool
  if _, v = currentToken(token.Key, "id2"); !v {
    t.Fatalf("Token %s should be valid", token.Key)
  }
  timeNow = func() time.Time { return time.Now().Add(time.Hour * 2) }
  if _, v := currentToken(token.Key, "id2"); v {
    t.Fatalf("Token %s should be invalid after timeout", token.Key)
  }

  timeNow = func() time.Time { return time.Now() }
  token = newToken(user2, "id3")
  if _, v = currentToken(token.Key, "id3"); !v {
    t.Fatalf("Token %s should be valid", token.Key)
  }
  timeNow = func() time.Time { return time.Now().Add(time.Hour * 2) }
  token.updateTimeout()
  if _, v := currentToken(token.Key, "id3"); !v {
    t.Fatalf("Token %s should be valid after timeout if refreshed", token.Key)
  }
  timeNow = func() time.Time { return time.Now().Add(time.Hour * 20) }
  token.updateTimeout()
  if _, v := currentToken(token.Key, "id3"); v {
    t.Fatalf("Token %s should be invalid after expiry even if refreshed", token.Key)
  }
}
