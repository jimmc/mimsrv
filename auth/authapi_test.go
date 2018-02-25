package auth

import (
  "net/http"
  "net/http/httptest"
  "testing"
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

  handlerResult := ""
  baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    handlerResult = "called"
  })
  wrappedHandler := h.RequireAuth(baseHandler)

  rr := httptest.NewRecorder()
  wrappedHandler.ServeHTTP(rr, req)
  if got, want := rr.Code, http.StatusUnauthorized; got != want {
    t.Errorf("request without auth: got status %d, want %d", got, want)
  }

  rr = httptest.NewRecorder()
  idstr := clientIdString(req)
  cookie := tokenCookie("user1", idstr)
  req.AddCookie(cookie)
  wrappedHandler.ServeHTTP(rr, req)
  if got, want := rr.Code, http.StatusOK; got != want {
    t.Errorf("request with auth: got status %d, want %d", got, want)
  }
}
