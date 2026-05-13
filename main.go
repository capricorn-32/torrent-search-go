package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/capricorn-32/torrent-search-go/configs"
	"github.com/capricorn-32/torrent-search-go/scrapers"
	"github.com/capricorn-32/torrent-search-go/web"
)

func main() {
	if err := configs.LoadScrapersConfig("configs/scrapers.json"); err != nil {
		log.Fatal(err)
	}

	// Check if running in CLI mode or web mode
	if len(os.Args) > 1 && os.Args[1] != "serve" {
		// CLI mode: search from command line
		query := strings.TrimSpace(strings.Join(os.Args[1:], " "))

		if query == "" {
			log.Fatal("search query cannot be empty")
		}

		scraper := scrapers.NyaaScraper{}
		torrents, err := scraper.Search(query)
		if err != nil {
			log.Fatal(err)
		}

		output, err := json.MarshalIndent(torrents, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(output))
		return
	}

	// Web mode: start HTTP server
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/search", web.SearchHandler)
	mux.HandleFunc("/api/health", web.HealthHandler)

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fs)

	// Start server
	port := ":8080"
	log.Printf("Starting server on http://localhost%s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}
