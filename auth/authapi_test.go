package auth

import (
  "net/http"
  "net/http/httptest"
  "testing"

  "github.com/jimmc/mimsrv/permissions"
  "github.com/jimmc/mimsrv/users"
)

func TestRequireAuth(t *testing.T) {
  h := NewHandler(&Config{
    Prefix: "/pre/",
    PasswordFilePath: "testdata/pw1.txt",
    MaxClockSkewSeconds: 2,
  })

  req, err := http.NewRequest("GET", "/api/list/d1", nil)
  if err != nil {
    t.Fatalf("error create auth list request: %v", err)
  }

  var reqUser *users.User
  baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    reqUser = CurrentUser(r)
  })
  wrappedHandler := h.RequireAuth(baseHandler)

  rr := httptest.NewRecorder()
  wrappedHandler.ServeHTTP(rr, req)
  if got, want := rr.Code, http.StatusUnauthorized; got != want {
    t.Errorf("request without auth: got status %d, want %d", got, want)
  }

  rr = httptest.NewRecorder()
  user := users.NewUser("user1", "cw1", nil)
  idstr := clientIdString(req)
  cookie := tokenCookie(user, idstr)
  req.AddCookie(cookie)
  reqUser = nil
  wrappedHandler.ServeHTTP(rr, req)
  if got, want := rr.Code, http.StatusOK; got != want {
    t.Errorf("request with auth: got status %d, want %d", got, want)
  }
  if reqUser == nil {
    t.Errorf("authenicated request should carry a current user")
  }
  if got, want := reqUser.Id(), user.Id(); got != want {
    t.Errorf("authenticated userid: got %s, want %s", got, want)
  }
  if got, want := reqUser.HasPermission(permissions.CanEdit), false; got != want {
    t.Errorf("permission for CanEdit: got %v, want %v", got, want)
  }

  req, err = http.NewRequest("GET", "/api/list/d1", nil)
  if err != nil {
    t.Fatalf("error create auth list request: %v", err)
  }
  rr = httptest.NewRecorder()
  user = users.NewUser("user1", "cw1", permissions.FromString("edit"))
  idstr = clientIdString(req)
  cookie = tokenCookie(user, idstr)
  req.AddCookie(cookie)
  reqUser = nil
  wrappedHandler.ServeHTTP(rr, req)
  if got, want := rr.Code, http.StatusOK; got != want {
    t.Errorf("request with auth: got status %d, want %d", got, want)
  }
  if reqUser == nil {
    t.Errorf("authenicated request should carry a current user")
  }
  if got, want := reqUser.HasPermission(permissions.CanEdit), true; got != want {
    t.Errorf("permission for CanEdit: got %v, want %v", got, want)
  }
}
