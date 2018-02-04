package main

import (
  "flag"
  "fmt"
  "log"
  "net/http"

  "github.com/jimmc/mimsrv/api"
  "github.com/jimmc/mimsrv/content"
)

type config struct {
  mimViewRoot string
  contentRoot string
}

func main() {
  config := &config{}
  flag.StringVar(&config.mimViewRoot, "mimviewroot", "", "location of mimview ui root (build/default)")
  flag.StringVar(&config.contentRoot, "contentroot", "", "root directory for content (photos)")
  flag.Parse()
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
  mux.Handle("/ui/", http.StripPrefix("/ui/", uiFileHandler))
  mux.Handle("/api/", api.NewHandler(&api.Config{
    Prefix: "/api/",
    ContentHandler: contentHandler,
  }))
  mux.HandleFunc("/", redirectToUi)

  fmt.Printf("mimsrv serving on port 8080\n")
  log.Fatal(http.ListenAndServe(":8080", mux))
}

func redirectToUi(w http.ResponseWriter, r *http.Request) {
  http.Redirect(w, r, "/ui/", http.StatusTemporaryRedirect)
}
