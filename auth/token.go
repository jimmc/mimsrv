package auth

import (
  "fmt"
  "math/rand"
  "time"

  "github.com/jimmc/mimsrv/users"
)

const (
  tokenTimeoutDuration = time.Duration(1) * time.Hour
  tokenExpirationDuration = time.Duration(10) * time.Hour
)

var (
  tokens map[string]*Token
)

type Token struct {
  Key string
  user *users.User
  idstr string
  timeout time.Time     // Time at which token is no longer valid if not refreshed
  expiry time.Time      // Time past which token can not be auto-refreshed
}

func initTokens() {
  tokens = make(map[string]*Token)
}

func newToken(user *users.User, idstr string) *Token {
  token := &Token{
    user: user,
    idstr: idstr,
    timeout: timeNow().Add(tokenTimeoutDuration),
    expiry: timeNow().Add(tokenExpirationDuration),
  }
  keynum := rand.Intn(1000000)
  token.Key = fmt.Sprintf("%06d", keynum)
  tokens[token.Key] = token
  return token
}

func currentToken(tokenKey, idstr string) (*Token, bool) {
  token := tokens[tokenKey]
  if token == nil {
    return nil, false
  }
  return token, token.isValid(idstr)
}

func (t *Token) isValid(idstr string) bool {
  if t.idstr != idstr {
    return false
  }
  if timeNow().After(t.timeout) {
    return false
  }
  return true
}

// updateTimeout reset the token timeout to be the timeout-duration
// from now, or the token expiry, whichever comes first.
func (t *Token) updateTimeout() {
  timeout := timeNow().Add(tokenTimeoutDuration)
  if timeout.After(t.expiry) {
    timeout = t.expiry
  }
  t.timeout = timeout
}

func (t *Token) User() *users.User {
  return t.user
}
