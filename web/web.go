package web

import (
	"fmt"
	"net/http"
)

type App struct{}

func (App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fmt.Fprint(w, "Hallo Welt!")
	case "/go":
		fmt.Fprint(w, "Du bist im Go-Pfad!")
	default:
		http.NotFound(w, r)
	}
}

func Start() {
	app := App{}
	http.ListenAndServe(":8080", app)
}
