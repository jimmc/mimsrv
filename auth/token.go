package auth

import (
  "fmt"
  "math/rand"
  "time"

  "github.com/jimmc/mimsrv/users"
)

const (
  tokenExpirationDuration = time.Duration(1) * time.Hour
)

var (
  tokens map[string]*Token
)

type Token struct {
  Key string
  user *users.User
  idstr string
  expiry time.Time
}

func initTokens() {
  tokens = make(map[string]*Token)
}

func newToken(user *users.User, idstr string) *Token {
  token := &Token{
    user: user,
    idstr: idstr,
    expiry: timeNow().Add(tokenExpirationDuration),
  }
  keynum := rand.Intn(1000000)
  token.Key = fmt.Sprintf("%06d", keynum)
  tokens[token.Key] = token
  return token
}

func isValidToken(tokenKey, idstr string) bool {
  token := tokens[tokenKey]
  if token == nil {
    return false
  }
  if token.idstr != idstr {
    return false
  }
  if timeNow().After(token.expiry) {
    return false
  }
  return true
}

func userFromToken(tokenKey string) *users.User {
  token := tokens[tokenKey]
  if token == nil {
    return nil
  }
  return token.user
}
