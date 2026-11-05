// Kanban-Board mit htmx und Go (nur Standardbibliothek).
// Start: go run . -> http://localhost:8080
package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxTitle = 120

var columns = []string{"Todo", "Doing", "Done"}

type card struct {
	ID    int
	Title string
	Col   string
	Prev  string // Ziel-Spalte links, leer wenn keine
	Next  string // Ziel-Spalte rechts, leer wenn keine
}

type colView struct {
	Name  string
	Count int
	Cards []card
}

type board struct {
	mu     sync.Mutex
	nextID int
	cards  []card
}

func validCol(c string) bool {
	for _, n := range columns {
		if n == c {
			return true
		}
	}
	return false
}

func (b *board) add(title string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	b.cards = append(b.cards, card{ID: b.nextID, Title: title, Col: columns[0]})
}

func (b *board) move(id int, col string) bool {
	if !validCol(col) {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := range b.cards {
		if b.cards[i].ID == id {
			b.cards[i].Col = col
			return true
		}
	}
	return false
}

func (b *board) remove(id int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := range b.cards {
		if b.cards[i].ID == id {
			b.cards = append(b.cards[:i], b.cards[i+1:]...)
			return true
		}
	}
	return false
}

// view gruppiert die Karten nach Spalte und berechnet die Nachbarspalten.
func (b *board) view() []colView {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]colView, len(columns))
	for i, name := range columns {
		out[i] = colView{Name: name}
	}
	for _, c := range b.cards {
		for i, name := range columns {
			if name != c.Col {
				continue
			}
			if i > 0 {
				c.Prev = columns[i-1]
			} else {
				c.Prev = ""
			}
			if i < len(columns)-1 {
				c.Next = columns[i+1]
			} else {
				c.Next = ""
			}
			out[i].Cards = append(out[i].Cards, c)
		}
	}
	for i := range out {
		out[i].Count = len(out[i].Cards)
	}
	return out
}

var boardTmpl = template.Must(template.New("board").Parse(`
<div id="board" class="board">
{{- range .}}
  <section class="column" data-col="{{.Name}}"
           ondragover="event.preventDefault(); this.classList.add('over')"
           ondragleave="this.classList.remove('over')"
           ondrop="dropCard(event, '{{.Name}}', this)">
    <header><h2>{{.Name}}</h2><span class="count">{{.Count}}</span></header>
    {{- range .Cards}}
    <article class="card" draggable="true" data-id="{{.ID}}"
             ondragstart="event.dataTransfer.setData('text/plain', '{{.ID}}')">
      <p>{{.Title}}</p>
      <div class="actions">
        {{- if .Prev}}
        <button hx-post="/cards/{{.ID}}/move" hx-vals='{"to":"{{.Prev}}"}' hx-target="#board" hx-swap="outerHTML" aria-label="Eine Spalte zurück">←</button>
        {{- end}}
        {{- if .Next}}
        <button hx-post="/cards/{{.ID}}/move" hx-vals='{"to":"{{.Next}}"}' hx-target="#board" hx-swap="outerHTML" aria-label="Eine Spalte weiter">→</button>
        {{- end}}
        <button class="ghost" hx-delete="/cards/{{.ID}}" hx-target="#board" hx-swap="outerHTML" aria-label="Karte löschen">×</button>
      </div>
    </article>
    {{- else}}
    <p class="empty">Keine Karten</p>
    {{- end}}
  </section>
{{- end}}
</div>
`))

func render(w http.ResponseWriter, cols []colView) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := boardTmpl.Execute(w, cols); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func cardID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ungültige ID", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func main() {
	b := &board{}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexPage))
	})

	mux.HandleFunc("GET /board", func(w http.ResponseWriter, r *http.Request) {
		render(w, b.view())
	})

	mux.HandleFunc("POST /cards", func(w http.ResponseWriter, r *http.Request) {
		title := strings.TrimSpace(r.FormValue("title"))
		if title == "" || len(title) > maxTitle {
			http.Error(w, "Titel fehlt oder ist zu lang", http.StatusBadRequest)
			return
		}
		b.add(title)
		render(w, b.view())
	})

	mux.HandleFunc("POST /cards/{id}/move", func(w http.ResponseWriter, r *http.Request) {
		id, ok := cardID(w, r)
		if !ok {
			return
		}
		if !b.move(id, r.FormValue("to")) {
			http.Error(w, "Karte oder Spalte nicht gefunden", http.StatusNotFound)
			return
		}
		render(w, b.view())
	})

	mux.HandleFunc("DELETE /cards/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := cardID(w, r)
		if !ok {
			return
		}
		if !b.remove(id) {
			http.Error(w, "Karte nicht gefunden", http.StatusNotFound)
			return
		}
		render(w, b.view())
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
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>htmx Kanban</title>
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
  <style>
    :root { color-scheme: dark; --bg:#0b0f1a; --panel:#131a2b; --card:#1d2640; --neon:#22d3ee; --pink:#f472b6; --text:#e2e8f0; }
    * { box-sizing: border-box; }
    body { margin:0; font-family: system-ui, sans-serif; background: radial-gradient(circle at 20% 0%, #1b1040, var(--bg) 60%); color: var(--text); min-height:100vh; padding:24px; }
    h1 { margin:0 0 16px; font-weight:600; letter-spacing:.04em; color: var(--neon); text-shadow: 0 0 12px rgba(34,211,238,.45); }
    form.new { display:flex; gap:8px; margin-bottom:20px; }
    input[name=title] { flex:1; padding:10px 12px; border-radius:8px; border:1px solid #334; background:var(--panel); color:var(--text); }
    button { padding:10px 14px; border-radius:8px; border:1px solid var(--neon); background:transparent; color:var(--neon); cursor:pointer; }
    button:hover { background: rgba(34,211,238,.12); }
    .board { display:grid; grid-template-columns: repeat(3, minmax(220px,1fr)); gap:16px; }
    .column { background:var(--panel); border-radius:14px; padding:12px; min-height:240px; border:1px solid #1f2a44; transition: box-shadow .15s, border-color .15s; }
    .column.over { border-color: var(--pink); box-shadow: 0 0 18px rgba(244,114,182,.35); }
    .column header { display:flex; justify-content:space-between; align-items:center; margin-bottom:10px; }
    .column h2 { font-size:1rem; margin:0; color: var(--pink); }
    .count { font-size:.8rem; opacity:.7; }
    .card { background:var(--card); border-radius:10px; padding:10px 12px; margin-bottom:10px; cursor:grab; border:1px solid #2a3760; }
    .card:active { cursor:grabbing; }
    .card p { margin:0 0 8px; }
    .actions { display:flex; gap:6px; justify-content:flex-end; }
    .actions button { padding:4px 8px; font-size:.8rem; }
    .ghost { border-color:#445 !important; color:#94a3b8 !important; }
    .empty { opacity:.45; font-style:italic; font-size:.9rem; }
    @media (max-width: 700px) { .board { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
  <h1>Kanban</h1>

  <form class="new" hx-post="/cards" hx-target="#board" hx-swap="outerHTML" hx-on::after-request="if(event.detail.successful) this.reset()">
    <input name="title" placeholder="Neue Karte" maxlength="120" required>
    <button type="submit">Hinzufügen</button>
  </form>

  <div id="board" hx-get="/board" hx-trigger="load" hx-swap="outerHTML"></div>

  <script>
    // Drag-and-drop: setzt die Ziel-Spalte per htmx-Request.
    function dropCard(event, col, el) {
      event.preventDefault();
      el.classList.remove('over');
      const id = event.dataTransfer.getData('text/plain');
      if (!id) return;
      htmx.ajax('POST', '/cards/' + id + '/move', {
        values: { to: col },
        target: '#board',
        swap: 'outerHTML'
      });
    }
  </script>
</body>
</html>
`
