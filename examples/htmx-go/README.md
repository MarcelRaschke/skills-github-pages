# htmx + Go Beispiel

Minimale Webanwendung mit htmx im Browser und einem Go-Server. Der Server nutzt nur die Standardbibliothek. Es gibt keine externen Go-Abhängigkeiten.

## Starten

    go run .

Danach im Browser `http://localhost:8080` öffnen.

## Endpunkte

| Methode | Pfad | Antwort |
|---|---|---|
| GET | `/` | HTML-Seite mit htmx-Attributen |
| GET | `/time` | Serverzeit als Fragment, alle 5 Sekunden aktualisiert |
| GET | `/items` | Liste der gespeicherten Einträge als `<li>`-Fragmente |
| POST | `/items` | Formularfeld `item` speichern, neues `<li>` zurückgeben |

## Hinweise

- Eingaben werden mit `html.EscapeString` ausgegeben, um XSS zu verhindern.
- Die Daten liegen nur im Speicher und gehen beim Neustart verloren.
- htmx wird über ein CDN geladen. Für Produktion die Datei lokal ausliefern.
- Requests ohne `item` erhalten den Status 400.
