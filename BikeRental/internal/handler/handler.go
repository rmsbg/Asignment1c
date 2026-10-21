// Package handler serves the initial browser interface.
package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"bikerental/internal/store"
	"bikerental/web"
)

type page struct {
	Path        string
	Title       string
	Description string
	Step        string
}

type pageData struct {
	Current   page
	Pages     []page
	Demo      bool
	Dashboard bool
	Counts    store.DashboardCounts
}

func New(state *store.Store) (http.Handler, error) {
	tmpl, err := template.ParseFS(web.Files, "page.gohtml")
	if err != nil {
		return nil, fmt.Errorf("Oberfläche laden: %w", err)
	}
	assets, err := fs.Sub(web.Files, "assets")
	if err != nil {
		return nil, fmt.Errorf("Assets laden: %w", err)
	}
	pages := []page{
		{Path: "/", Title: "Dashboard", Description: "Alles für den Verleihtag an einem Ort.", Step: "Übersicht"},
		{Path: "/fahrraeder", Title: "Fahrräder verwalten", Description: "Der richtige Überblick über jedes Fahrrad.", Step: "Inventar"},
		{Path: "/personen", Title: "Personen verwalten", Description: "Kunden erfassen. Den Überblick behalten.", Step: "Kunden"},
		{Path: "/tarife", Title: "Tarife verwalten", Description: "Ein klarer Preis für jeden Verleihzeitraum.", Step: "Tarife"},
		{Path: "/ausleihe", Title: "Neue Ausleihe", Description: "Vom verfügbaren Fahrrad zur nächsten Fahrt.", Step: "Verleih"},
	}
	mux := http.NewServeMux()
	registerAPI(mux, state)
	registerRentals(mux, state)
	mux.HandleFunc("GET /api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			return
		}
		writeJSON(w, http.StatusOK, state.Dashboard())
	})
	for _, current := range pages {
		pattern := "GET " + current.Path
		if current.Path == "/" {
			pattern = "GET /{$}"
		}
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			data := pageData{Current: current, Pages: pages, Demo: state.IsDemo(), Dashboard: current.Path == "/"}
			if data.Dashboard {
				data.Counts = state.Dashboard()
			}
			var body bytes.Buffer
			if err := tmpl.Execute(&body, data); err != nil {
				log.Printf("Oberfläche rendern: %v", err)
				http.Error(w, "Die Ansicht konnte nicht geladen werden.", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			if r.Method != http.MethodHead {
				_, _ = w.Write(body.Bytes())
			}
		})
	}
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServerFS(assets)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	}), nil
}
