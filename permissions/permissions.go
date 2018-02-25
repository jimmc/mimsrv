package permissions

import (
  "strings"
)

const permSepChar = " "

type Permission int
const (
  CanEdit Permission = iota +1
)

type Permissions struct {
  perms map[Permission]bool
}

func FromString(permstr string) *Permissions {
  permstr = strings.TrimSpace(permstr)
  p := &Permissions{
    perms: make(map[Permission]bool),
  }
  pp := strings.Split(permstr, permSepChar)
  for _, pstr := range pp {
    if pstr != "" {
      perm := permFromString(pstr)
      p.perms[perm] = true;
    }
  }
  return p
}

func (p *Permissions) ToString() string {
  s := ""
  sep := ""
  for perm, _ := range p.perms {
    pstr := permToString(perm)
    s = s + sep + pstr
    sep = permSepChar
  }
  return s
}

func (p *Permissions) HasPermission(perm Permission) bool {
  return p.perms[perm]
}

func permFromString(s string) Permission {
  if s == "edit" {
    return CanEdit
  }
  return 0      // No valid permission string found
}

func permToString(perm Permission) string {
  if perm == CanEdit {
    return "edit"
  }
  return ""
}
