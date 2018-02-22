package main

import (
  "bytes"
  "io/ioutil"
  "os"
  "testing"

  "github.com/jimmc/mimsrv/auth"
)

const (
  maxClockSkewSeconds = 2
)

func TestPasswordFile(t *testing.T) {
  pwfile := "testdata/password.tmp"
  bakfile := pwfile + "~"
  defer os.Remove(pwfile)
  defer os.Remove(bakfile)
  _ = os.Remove(pwfile) // Pre-clean

  authHandler := auth.NewHandler(&auth.Config{
    Prefix: "/auth/",
    PasswordFilePath: pwfile,
    MaxClockSkewSeconds: maxClockSkewSeconds,
  })

  err := authHandler.CreatePasswordFile()
  if err != nil {
    t.Errorf("Should have created password file %s: %v", pwfile, err)
  }
  err = authHandler.CreatePasswordFile()
  if err == nil {
    t.Errorf("CreatePasswordFile should fail when file exists")
  }

  err = authHandler.UpdatePassword("user1", "pw1")
  if err != nil {
    t.Fatalf("Error updating password file %s: %v", pwfile, err)
  }

  pwgot, err := ioutil.ReadFile(pwfile)
  if err != nil {
    t.Fatalf("Failed to read temp password file: %v", err)
  }
  pwwant, err := ioutil.ReadFile("testdata/password-demo.txt")
  if err != nil {
    t.Fatalf("Failed to read reference password file: %v", err)
  }
  if !bytes.Equal(pwgot, pwwant) {
    t.Errorf("password file contents don't match, got '%s', want '%s'", pwgot, pwwant)
  }
}
