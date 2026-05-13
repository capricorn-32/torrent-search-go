package scrapers

import (
	"strconv"
	"strings"

	"github.com/capricorn-32/torrent-search-go/models"
)

type TorrentScraper interface {
	Search(query string) ([]models.Torrent, error)
	Name() string
}

func parseInt(s string) int {
	num, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return num
}
