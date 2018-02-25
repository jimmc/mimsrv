package permissions

import (
  "testing"
)

func TestFromString(t *testing.T) {
  p := FromString("")
  if got, want := len(p.perms), 0; got != want {
    t.Errorf("Number of permissions in empty string: got %d, want %d", got, want)
  }
  if p.HasPermission(CanEdit) {
    t.Errorf("empty string should not give CanEdit permission")
  }

  p = FromString("edit")
  if got, want := len(p.perms), 1; got != want {
    t.Errorf("Number of permissions in 'edit' string: got %d, want %d", got, want)
  }
  if !p.HasPermission(CanEdit) {
    t.Errorf("'edit' string fails to give CanEdit permission")
  }
}
