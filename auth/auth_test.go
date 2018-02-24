package auth

import (
  "io/ioutil"
  "os"
  "testing"
  "time"
)

var (
  testConfig = &Config{
    Prefix: "/pre/",
    PasswordFilePath: "/tmp/mimsrv-passwd-test.txt",
    MaxClockSkewSeconds: 2,
  }
)

func TestPasswordFile(t *testing.T) {
  pf, err := ioutil.TempFile("", "mimsrv-test")
  if err != nil {
    t.Fatalf("failed to create temp password file")
  }
  defer os.Remove(pf.Name())    // clean up
  c := &Config{
    Prefix: "/pre/",
    PasswordFilePath: pf.Name(),
    MaxClockSkewSeconds: 2,
  }
  h := NewHandler(c)
  err = h.loadPasswordFile()
  if err != nil {
    t.Errorf("failed to load empty password file")
  }
  if got, want := h.users.UserCount(), 0; got != want {
    t.Errorf("empty tmp password file got %d records, want %d", got, want)
  }
  err = pf.Close()
  if err != nil {
    t.Errorf("error closing tmp password file")
  }

  err = h.CreatePasswordFile()
  if err == nil {
    t.Errorf("attempting to create existing password file should fail")
  }
  err = os.Remove(pf.Name())
  if err != nil {
    t.Errorf("failed to remove tmp password file: %v", err)
  }
  err = h.CreatePasswordFile()
  if err != nil {
    t.Errorf("failed to create password file")
  }

  cw := h.getCryptword("user1")
  if cw != "" {
    t.Errorf("user1 should have no cryptword before being set")
  }
  err = h.UpdatePassword("user1", "abcd")
  if err != nil {
    t.Errorf("failed to update password")
  }
  err = h.loadPasswordFile()
  if err != nil {
    t.Errorf("failed to load password file after updating")
  }
  if got, want := h.getCryptword("user1"), h.generateCryptword("user1", "abcd"); got != want {
    t.Errorf("user cryptword after saving: got %s, want %s", got, want)
  }
}

func TestCryptword(t *testing.T) {
  h := NewHandler(testConfig)
  cw := h.getCryptword("user1")
  if cw != "" {
    t.Errorf("cryptword for unknown user should be blank")
  }
  h.setCryptword("user1", "abcdef")
  cw = h.getCryptword("user1")
  if cw != "abcdef" {
    t.Errorf("cryptword should be equal to what was previously set")
  }
  h.setCryptword("user1", "ghi")
  cw = h.getCryptword("user1")
  if cw != "ghi" {
    t.Errorf("cryptword should be equal to new value")
  }
}

func TestGenerateNonceAtTime(t *testing.T) {
  h := NewHandler(testConfig)
  t0 := int64(1000000)
  t1 := t0 + 1
  nonce0 := h.generateNonceAtTime("user1", t0)
  if nonce0 == "" {
    t.Errorf("nonce should not be empty")
  }
  nonce1 := h.generateNonceAtTime("user1", t1)
  if nonce0 == nonce1 {
    t.Errorf("nonces generated at different times should be different")
  }
}

func TestNonceIsValidAtTime(t *testing.T) {
  h := NewHandler(testConfig)
  t0 := int64(1000000)
  t1 := t0 + 1
  nonce := h.generateNonceAtTime("user1", t0)
  if !h.nonceIsValidAtTime("user1", nonce, t0) {
    t.Errorf("nonce should be valid at same time as generated")
  }
  if h.nonceIsValidAtTime("user1", nonce, t1) {
    t.Errorf("nonce should not be valid at different time as generated")
  }
}

func TestNonceIsValidNow(t *testing.T) {
  h := NewHandler(testConfig)
  t0 := int64(1000000)
  t1 := t0
  timeNow = func() time.Time { return time.Unix(t1, 0) }
  nonce := h.generateNonceAtTime("user1", t1)
  if !h.nonceIsValidNow("user1", nonce, t0) {
    t.Errorf("nonce should be valid at same time as generated")
  }
  t1 = t0 + 1
  if !h.nonceIsValidNow("user1", nonce, t0) {
    t.Errorf("nonce should be valid at earlier time within skew")
  }
  t1 = t0 - 1
  if !h.nonceIsValidNow("user1", nonce, t0) {
    t.Errorf("nonce should be valid at later time within skew")
  }
  t1 = t0 - int64(testConfig.MaxClockSkewSeconds) - 1
  if h.nonceIsValidNow("user1", nonce, t0) {
    t.Errorf("nonce should not be valid at earlier time outside skew")
  }
  t1 = t0 + int64(testConfig.MaxClockSkewSeconds) + 1
  if h.nonceIsValidNow("user1", nonce, t0) {
    t.Errorf("nonce should not be valid at later time outside skew")
  }
}

func TestSha256sum(t *testing.T) {
  input := "abc-def"
  want := "abe70a7e804fcd4069cdee57873899c152b2f1eace1f2fd89b1a6e9b862481b9"
  if got := sha256sum(input); got != want {
    t.Errorf("sha256sum got %s want %s", got, want)
  }
}
