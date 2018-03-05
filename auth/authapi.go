package auth

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "strconv"
  "time"

  "github.com/jimmc/mimsrv/permissions"
  "github.com/jimmc/mimsrv/users"
)

const (
  tokenCookieName = "MIMSRV_TOKEN"
)

type LoginStatus struct {
  LoggedIn bool
  Permissions string
}

type authKey int
const (
  ctxUserKey = iota + 1
)

func (h *Handler) initApiHandler() {
  mux := http.NewServeMux()
  mux.HandleFunc(h.apiPrefix("login"), h.login)
  mux.HandleFunc(h.apiPrefix("logout"), h.logout)
  mux.HandleFunc(h.apiPrefix("status"), h.status)
  h.ApiHandler = mux
}

func (h *Handler) RequireAuth(httpHandler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
    tokenKey := cookieValue(r, tokenCookieName)
    idstr := clientIdString(r)
    if token, valid := currentToken(tokenKey, idstr); valid {
      token.updateTimeout()
      http.SetCookie(w, token.cookie()) // Set the renewed cookie
      user := token.User()
      mimRequest := requestWithContextUser(r, user)
      httpHandler.ServeHTTP(w, mimRequest)
    } else {
      // No token, or token is not valid
      http.Error(w, "Invalid token", http.StatusUnauthorized)
    }
  })
}

func requestWithContextUser(r *http.Request, user *users.User) *http.Request {
  mimContext := context.WithValue(r.Context(), ctxUserKey, user)
  return r.WithContext(mimContext)
}

func CurrentUser(r *http.Request) *users.User {
  v := r.Context().Value(ctxUserKey)
  if v == nil {
    return nil
  }
  return v.(*users.User)
}

func CurrentUserHasPermission(r *http.Request, perm permissions.Permission) bool {
  user := CurrentUser(r)
  if user == nil {
    return false
  }
  return user.HasPermission(perm)
}

func (h *Handler) apiPrefix(s string) string {
  return fmt.Sprintf("%s%s/", h.config.Prefix, s)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
  userid := r.FormValue("userid")
  nonce := r.FormValue("nonce")
  timestr := r.FormValue("time")
  seconds, err := strconv.ParseInt(timestr, 10, 64)
  if err != nil {
    log.Printf("Error converting time string '%s': %v\n", timestr, err)
    seconds = 0
  }

  user := h.users.User(userid)
  if user != nil && h.nonceIsValidNow(userid, nonce, seconds) {
    // OK to log in; generate a bearer token and put in a cookie
    idstr := clientIdString(r)
    http.SetCookie(w, tokenCookie(user, idstr))
  } else {
    http.Error(w, "Invalid userid or nonce", http.StatusUnauthorized)
    return
  }

  result := &LoginStatus{
    LoggedIn: true,
    Permissions: user.PermissionsString(),
  }
  b, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to marshall login status: %v", err), http.StatusInternalServerError)
    return
  }
  w.WriteHeader(http.StatusOK)
  w.Write(b)
}

func tokenCookie(user *users.User, idstr string) *http.Cookie {
  token := newToken(user, idstr)
  return token.cookie()
}

func (t *Token) cookie() *http.Cookie {
  return &http.Cookie{
    Name: tokenCookieName,
    Path: "/",
    Value: t.Key,
    Expires: t.timeout,
    HttpOnly: true,
  }
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
  // Clear our token cookie
  tokenCookie := &http.Cookie{
    Name: tokenCookieName,
    Path: "/",
    Value: "",
    Expires: time.Now().AddDate(-1, 0, 0),
  }
  http.SetCookie(w, tokenCookie)
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status": "ok"}`))
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
  tokenKey := cookieValue(r, tokenCookieName)
  idstr := clientIdString(r)
  token, loggedIn := currentToken(tokenKey, idstr)
  result := &LoginStatus{
    LoggedIn: loggedIn,
  }
  if loggedIn {
    token.updateTimeout()
    http.SetCookie(w, token.cookie()) // Set the renewed cookie
    result.Permissions = token.User().PermissionsString()
  }

  b, err := json.MarshalIndent(result, "", "  ")
  if err != nil {
    http.Error(w, fmt.Sprintf("Failed to marshall login status: %v", err), http.StatusInternalServerError)
    return
  }
  w.WriteHeader(http.StatusOK)
  w.Write(b)
}

func clientIdString(r *http.Request) string {
  return r.UserAgent()
}

func cookieValue(r *http.Request, cookieName string) string {
  cookie, err := r.Cookie(cookieName)
  if err != nil {
    return ""
  }
  return cookie.Value
}
