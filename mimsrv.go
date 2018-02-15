package main

import (
  "flag"
  "fmt"
  "log"
  "net/http"
  "os"
  "strconv"

  "github.com/jimmc/mimsrv/api"
  "github.com/jimmc/mimsrv/auth"
  "github.com/jimmc/mimsrv/content"
)

type config struct {
  port int
  mimViewRoot string
  contentRoot string
  passwordFilePath string
  maxClockSkewSeconds int
}

func main() {
  config := &config{}

  flag.IntVar(&config.port, "port", 8080, "port on which to listen for connections")
  flag.StringVar(&config.mimViewRoot, "mimviewroot", "", "location of mimview ui root (build/default)")
  flag.StringVar(&config.contentRoot, "contentroot", "", "root directory for content (photos)")
  flag.StringVar(&config.passwordFilePath, "passwordfile", "", "location of password file")
  flag.IntVar(&config.maxClockSkewSeconds, "maxclockskewseconds", 2, "max allowed skew between client and server")

  createPasswordP := flag.Bool("createPasswordFile", false, "create an empty password file")
  updatePasswordP := flag.String("updatePassword", "", "update password for named user")

  flag.Parse()

  if config.passwordFilePath == "" {
    log.Fatal("--passwordfile is required")
  }
  authHandler := auth.NewHandler(&auth.Config{
    Prefix: "/auth/",
    PasswordFilePath: config.passwordFilePath,
    MaxClockSkewSeconds: config.maxClockSkewSeconds,
  })
  if (*createPasswordP) {
    err := authHandler.CreatePasswordFile()
    if err != nil {
      fmt.Printf("Error creating password file: %v\n", err)
      os.Exit(1)
    }
    fmt.Printf("Password file created at %s\n", config.passwordFilePath)
    os.Exit(0)
  }
  if (*updatePasswordP != "") {
    err := authHandler.UpdateUserPassword(*updatePasswordP)
    if err != nil {
      fmt.Printf("Error updating password for %s: %v\n", *updatePasswordP, err)
      os.Exit(1)
    }
    fmt.Printf("Password updated for %s\n", *updatePasswordP)
    os.Exit(0)
  }

  if config.mimViewRoot == "" {
    log.Fatal("--mimviewroot is required")
  }
  if config.contentRoot == "" {
    log.Fatal("--contentroot is required")
  }

  mux := http.NewServeMux()

  contentHandler := content.NewHandler(&content.Config{
    ContentRoot: config.contentRoot,
  })
  uiFileHandler := http.FileServer(http.Dir(config.mimViewRoot))
  apiHandler := api.NewHandler(&api.Config{
    Prefix: "/api/",
    ContentHandler: contentHandler,
  })
  mux.Handle("/ui/", http.StripPrefix("/ui/", uiFileHandler))
  mux.Handle("/api/", authHandler.RequireAuth(apiHandler))
  mux.Handle("/auth/", authHandler.ApiHandler)
  mux.HandleFunc("/", redirectToUi)

  fmt.Printf("mimsrv serving on port %v\n", config.port)
  log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.port), mux))
}

func redirectToUi(w http.ResponseWriter, r *http.Request) {
  http.Redirect(w, r, "/ui/", http.StatusTemporaryRedirect)
}
