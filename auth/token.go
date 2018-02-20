package auth

import (
  "fmt"
  "math/rand"
  "time"
)

const (
  tokenExpirationDuration = time.Duration(1) * time.Hour
)

var (
  tokens map[string]*Token
)

type Token struct {
  Key string
  userid string
  idstr string
  expiry time.Time
}

func initTokens() {
  tokens = make(map[string]*Token)
}

func newToken(userid, idstr string) *Token {
  token := &Token{
    userid: userid,
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
