package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"sync"
	"time"
)

type store struct {
	mu    sync.Mutex
	items []string
}

func (s *store) add(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, v)
}

func (s *store) list() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.items...)
}

func main() {
	s := &store{}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexPage)
	})

	mux.HandleFunc("GET /time", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<span>%s</span>", time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		v := r.FormValue("item")
		if v == "" {
			http.Error(w, "item fehlt", http.StatusBadRequest)
			return
		}
		s.add(v)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(v))
	})

	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		for _, it := range s.list() {
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(it))
		}
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("läuft auf http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}

const indexPage = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <title>htmx + Go</title>
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body>
  <h1>htmx + Go</h1>
  <p>Serverzeit: <span hx-get="/time" hx-trigger="load, every 5s" hx-swap="innerHTML">lädt…</span></p>

  <form hx-post="/items" hx-target="#items" hx-swap="beforeend" hx-on::after-request="if(event.detail.successful) this.reset()">
    <input name="item" placeholder="Eintrag" required>
    <button type="submit">Hinzufügen</button>
  </form>

  <ul id="items" hx-get="/items" hx-trigger="load"></ul>
</body>
</html>
`
