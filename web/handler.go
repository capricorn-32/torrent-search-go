package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/capricorn-32/torrent-search-go/scrapers"
)

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Results []interface{} `json:"results"`
	Error   string        `json:"error,omitempty"`
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SearchResponse{Error: "method not allowed"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SearchResponse{Error: "query parameter 'q' is required"})
		return
	}

	scraper := scrapers.NyaaScraper{}
	torrents, err := scraper.Search(query)
	if err != nil {
		log.Printf("Search error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SearchResponse{Error: err.Error()})
		return
	}

	results := make([]interface{}, len(torrents))
	for i, t := range torrents {
		results[i] = t
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SearchResponse{Results: results})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
