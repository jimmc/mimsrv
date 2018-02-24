package users

import (
  "bytes"
  "io/ioutil"
  "os"
  "testing"
)

func TestEmpty(t *testing.T) {
  m := Empty();
  if got, want := m.UserCount(), 0; got != want {
    t.Errorf("user count for initial Empty: got %d, want %d", got, want)
  }
  m.addUser("user1", "crypt1")
  if got, want := m.UserCount(), 1; got != want {
    t.Errorf("user count after adding a user: got %d, want %d", got, want)
  }
}

func TestLoadSaveFile(t *testing.T) {
  pwfile := "testdata/pw1.txt"
  m, err := LoadFile(pwfile)
  if err != nil {
    t.Fatalf("failed to load password file %s: %v", pwfile, err)
  }

  if got, want := m.UserCount(), 2; got != want {
    t.Fatalf("user count in password file %s: got %d, want %d", pwfile, got, want)
  }
  if got, want := m.Cryptword("user1"), "d761bfe5ffda189a8f1c2212c5fb3fe65274a070d0b1c4f4ec6c2c020db5f22b";
      got != want {
    t.Errorf("cryptword for user1: got %s, want %s", got, want)
  }

  m.addUser("user3", "cw3")
  m.SetCryptword("user2", "cw2")

  pwsavefile := "testdata/tmp-pw-save.txt"
  pwsavebakfile := "testdata/tmp-pw-save.txt~"
  os.Remove(pwsavefile)
  defer os.Remove(pwsavefile)
  defer os.Remove(pwsavebakfile)
  // Pre-create the old file to be moved when the new is saved
  oldContents := []byte("old pw file")
  err = ioutil.WriteFile(pwsavefile, oldContents, 0644)
  if err != nil {
    t.Fatalf("failed to precreate saved password file %s: %v:", pwsavefile, err)
  }
  err = m.SaveFile(pwsavefile)
  if err != nil {
    t.Fatalf("error saving new password file %s: %v", pwsavefile, err)
  }

  // Make sure we renamed the old file as a backup
  pwgot, err := ioutil.ReadFile(pwsavebakfile)
  if err != nil {
    t.Errorf("failed to load save-backup password file %s: %v", pwsavebakfile, err)
  }
  if !bytes.Equal(pwgot, oldContents) {
    t.Errorf("save password file contents: got %s, want %s", pwgot, oldContents)
  }

  // Make sure the new password file is correct
  pwgolden := "testdata/pw2-golden.txt"
  pwgot, err = ioutil.ReadFile(pwsavefile)
  if err != nil {
    t.Fatalf("Failed to read saved password file %s: %v", pwsavefile, err)
  }
  pwwant, err := ioutil.ReadFile(pwgolden)
  if err != nil {
    t.Fatalf("Failed to read reference password file %s: %v", pwgolden, err)
  }
  if !bytes.Equal(pwgot, pwwant) {
    t.Errorf("password file contents don't match, got '%s', want '%s'", pwgot, pwwant)
  }
}
