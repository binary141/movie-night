package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type OMDB struct {
	key    string
	client *http.Client
}

func NewOMDB(key string) *OMDB {
	if key == "" {
		return nil
	}
	return &OMDB{key: key, client: &http.Client{Timeout: 5 * time.Second}}
}

type SearchResult struct {
	Title  string
	ImdbID string
	Year   string
	Poster string
}

type TitleSearchResult struct {
	ImdbID     string
	ImdbRating string
	Runtime    string
	Plot       string
	Genre      string
	Director   string
	Rated      string
}

func (o *OMDB) Search(ctx context.Context, query string) ([]SearchResult, error) {
	u := "https://www.omdbapi.com/?type=movie&apikey=" + url.QueryEscape(o.key) +
		"&s=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body struct {
		Search []struct {
			Title  string `json:"Title"`
			ImdbID string `json:"imdbID"`
			Year   string `json:"Year"`
			Poster string `json:"Poster"`
		} `json:"Search"`
		Response string `json:"Response"`
		Error    string `json:"Error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.Response != "True" {
		// "Movie not found!" and "Too many results." are normal outcomes, not errors.
		if body.Error == "Movie not found!" || body.Error == "Too many results." {
			return nil, nil
		}
		return nil, fmt.Errorf("omdb: %s", body.Error)
	}

	results := make([]SearchResult, 0, len(body.Search))
	for _, r := range body.Search {
		poster := r.Poster
		if poster == "N/A" {
			poster = ""
		}
		log.Printf("%+v", r.ImdbID)
		results = append(results, SearchResult{Title: r.Title, ImdbID: r.ImdbID, Year: r.Year, Poster: poster})
	}
	return results, nil
}

// SearchTitle looks up full details for a single movie, preferring an exact
// imdbID match and falling back to a title lookup when no id is known (e.g.
// a manually typed title with no search result behind it).
func (o *OMDB) SearchTitle(ctx context.Context, imdbID, title string) (*TitleSearchResult, error) {
	u := "https://www.omdbapi.com/?type=movie&apikey=" + url.QueryEscape(o.key)
	if imdbID != "" {
		u += "&i=" + url.QueryEscape(imdbID)
	} else {
		u += "&t=" + url.QueryEscape(title)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body struct {
		ImdbID     string `json:"imdbID"`
		ImdbRating string `json:"imdbRating"`
		Runtime    string `json:"Runtime"`
		Rated      string `json:"Rated"`
		Genre      string `json:"Genre"`
		Director   string `json:"Director"`
		Plot       string `json:"Plot"`
		Response   string `json:"Response"`
		Error      string `json:"Error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.Response != "True" {
		// "Movie not found!" is a normal outcome, not an error.
		if body.Error == "Movie not found!" {
			return nil, nil
		}
		return nil, fmt.Errorf("omdb: %s", body.Error)
	}

	na := func(s string) string {
		if s == "N/A" {
			return ""
		}
		return s
	}
	return &TitleSearchResult{
		ImdbID:     body.ImdbID,
		ImdbRating: na(body.ImdbRating),
		Runtime:    na(body.Runtime),
		Rated:      na(body.Rated),
		Genre:      na(body.Genre),
		Director:   na(body.Director),
		Plot:       na(body.Plot),
	}, nil
}
