package users

import (
  "bufio"
  "encoding/csv"
  "fmt"
  "log"
  "os"
)

type User struct {
  userid string
  cryptword string
}

type Users struct {
  records [][]string
  users map[string]*User
}

func Empty() *Users {
  return &Users{
    records: make([][]string, 0),
    users: make(map[string]*User),
  }
}

func LoadFile(filename string) (*Users, error) {
  f, err := os.Open(filename)
  if err != nil {
    return nil, fmt.Errorf("error opening password file %s: %v", filename, err)
  }
  r := csv.NewReader(bufio.NewReader(f))

  records, err := r.ReadAll()
  if err != nil {
    return nil, fmt.Errorf("error loading password file %s: %v", filename, err)
  }

  users := users(records)
  return &Users{
    records: records,
    users: users,
  }, nil
}

func users(records [][]string) map[string]*User {
  users := make(map[string]*User)
  for _, record := range records {
    userid := record[0]
    user := &User{
      userid: userid,
      cryptword: record[1],
    }
    users[userid] = user
  }
  return users
}

func (m *Users) SaveFile(filename string) error {
  newFilePath := filename + ".new"
  f, err := os.Create(newFilePath)
  if err != nil {
    return fmt.Errorf("error creating new password file %s: %v", newFilePath, err)
  }
  w := csv.NewWriter(bufio.NewWriter(f))
  err = w.WriteAll(m.records)
  if err != nil {
    return fmt.Errorf("error writing new password file %s: %v", newFilePath, err)
  }
  w.Flush()
  f.Close()

  backupFilePath := filename + "~"
  err = os.Rename(filename, backupFilePath)
  if err != nil {
    return fmt.Errorf("error moving old file to backup path %s: %v", backupFilePath, err)
  }
  err = os.Rename(newFilePath, filename)
  if err != nil {
    return fmt.Errorf("error moving new file %s to become active file: %v", newFilePath, err)
  }

  return nil
}

func (m *Users) UserCount() int {
  return len(m.records)
}

func (m *Users) addUser(userid, cryptword string) {
  m.addUserOnly(userid, cryptword)
  m.addRecord(userid, cryptword)
}

func (m *Users) addUserOnly(userid, cryptword string) {
  user := &User{
    userid: userid,
    cryptword: cryptword,
  }
  m.users[userid] = user
}

func (m *Users) addRecord(userid, cryptword string) {
  for r, record := range m.records {
    if record[0] == userid {
      // This user already exists, update the existing record
      m.records[r][1] = cryptword
      return
    }
  }
  record := []string{userid, cryptword}
  m.records = append(m.records, record)
}

func (m *Users) User(userid string) *User {
  return m.users[userid]
}

func (m *Users) SetCryptword(userid, cryptword string) {
  for r, record := range(m.records) {
    if record[0] == userid {
      m.records[r][1] = cryptword
      user := m.User(userid)
      if user == nil {  // should never happen
        log.Printf("Error setting cryptword, user %s is in records but not users", userid)
        return
      }
      user.SetCryptword(cryptword)
      return
    }
  }
  record := []string{userid, cryptword}
  m.records = append(m.records, record)
  m.addUser(userid, cryptword)
}

func (m *Users) Cryptword(userid string) string {
  user := m.User(userid)
  if user == nil {
    return ""
  }
  return user.Cryptword()
}

func (u *User) Cryptword() string {
  return u.cryptword
}

func (u *User) SetCryptword(cryptword string) {
  u.cryptword = cryptword
}
