# htmx Kanban Board

Kanban-Board mit drei Spalten (Todo, Doing, Done). Der Server nutzt nur die Go-Standardbibliothek. htmx übernimmt die Kommunikation, die Oberfläche bleibt schlank.

## Starten

    go run .

Danach `http://localhost:8080` öffnen.

## Funktionen

- Karte anlegen über das Formular oben (maximal 120 Zeichen)
- Karte verschieben per Drag-and-drop auf eine Spalte
- Karte verschieben per Buttons ← und →
- Karte löschen über ×

## Endpunkte

| Methode | Pfad | Zweck |
|---|---|---|
| GET | `/` | Seite mit htmx und Styles |
| GET | `/board` | Board als HTML-Fragment |
| POST | `/cards` | Formularfeld `title`, neue Karte in Todo |
| POST | `/cards/{id}/move` | Feld `to` mit Zielspalte |
| DELETE | `/cards/{id}` | Karte löschen |

Jede Änderung antwortet mit dem vollständigen Board als Fragment. htmx ersetzt damit das alte Board per `outerHTML`.

## Hinweise

- Daten liegen nur im Speicher und sind nach einem Neustart weg.
- Ein `sync.Mutex` schützt den Zustand bei parallelen Requests.
- Zielspalten werden serverseitig geprüft. Unbekannte Spalten werden abgelehnt.
- Titel werden getrimmt und dürfen nicht leer sein. Alle Ausgaben sind escaped.
- Die Live-Aktualisierung zwischen mehreren Browsern fehlt. Dafür wäre ein Polling-Wrapper oder Server-Sent Events nötig.
- htmx wird über ein CDN geladen. Für Produktion lokal ausliefern.
