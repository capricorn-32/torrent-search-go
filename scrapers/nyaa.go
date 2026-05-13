package scrapers

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/capricorn-32/torrent-search-go/configs"
	"github.com/capricorn-32/torrent-search-go/models"
	"github.com/gocolly/colly/v2"
)

type NyaaScraper struct{}

func (s NyaaScraper) Name() string {
	return "NyaaSI"
}

func (s NyaaScraper) Search(query string) ([]models.Torrent, error) {
	cfg, exists := configs.GetScraperConfig("nyaa")
	if !exists {
		return nil, fmt.Errorf("scraper config for nyaa is not loaded")
	}

	var torrents []models.Torrent

	c := colly.NewCollector(
		colly.AllowedDomains(cfg.Domains...),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"),
	)
	c.AllowURLRevisit = true
	c.WithTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		DisableKeepAlives:     false,
		MaxIdleConnsPerHost:   10,
		DisableCompression:    false,
	})
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Referer", "https://nyaa.si/")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "same-origin")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("sec-ch-ua", `"Chromium";v="124", "Google Chrome";v="124", ";Not A Brand";v="99"`)
		r.Headers.Set("sec-ch-ua-mobile", "?0")
		r.Headers.Set("sec-ch-ua-platform", `"macOS"`)
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       1 * time.Second,
	})

	c.OnHTML("table.torrent-list tbody tr", func(e *colly.HTMLElement) {
		name := strings.TrimSpace(e.ChildText("td:nth-child(2) a:last-child"))

		quality := models.ExtractQuality(name)
		language := models.DetectLanguage(name)

		if strings.Contains(strings.ToLower(name), "raw") {
			language = "japanese"
		}

		torrent := models.Torrent{
			Name:       name,
			URL:        cfg.BaseURL + e.ChildAttr("td:nth-child(2) a:last-child", "href"),
			Size:       e.ChildText("td:nth-child(4)"),
			UploadDate: e.ChildText("td:nth-child(5)"),
			Quality:    quality,
			Source:     s.Name(),
			Language:   language,
			MagnetLink: e.ChildAttr("td:nth-child(3) a:nth-child(2)", "href"),
		}

		seedersStr := e.ChildText("td:nth-child(6)")
		seeders, err := strconv.Atoi(seedersStr)
		if err == nil {
			torrent.Seeders = seeders
		}

		leechersStr := e.ChildText("td:nth-child(7)")
		leechers, err := strconv.Atoi(leechersStr)
		if err == nil {
			torrent.Leechers = leechers
		}

		torrents = append(torrents, torrent)
	})

	formattedQuery := strings.ReplaceAll(query, " ", "+")
	searchURL, err := configs.BuildSearchURL("nyaa", formattedQuery, "+")

	if err != nil {
		return nil, err
	}

	var visitErr error
	for attempt := 1; attempt <= 3; attempt++ {
		visitErr = c.Visit(searchURL)
		if visitErr == nil {
			break
		}

		if attempt < 3 && isTransientNetworkError(visitErr) {
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		return nil, fmt.Errorf("error visiting %s: %w", searchURL, visitErr)
	}

	c.Wait()
	return torrents, nil
}

func isTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}

	if ne, ok := err.(net.Error); ok && (ne.Timeout() || ne.Temporary()) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "connection reset") ||
		strings.Contains(message, "reset by peer") ||
		strings.Contains(message, "timeout") ||
		strings.Contains(message, "temporary") ||
		strings.Contains(message, "tls handshake timeout")
}
