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
	ImdbRating string
	Runtime    string
	Plot       string
	Genre      string
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

func (o *OMDB) SearchTitle(ctx context.Context, title string) ([]TitleSearchResult, error) {
	u := "https://www.omdbapi.com/?type=movie&apikey=" + url.QueryEscape(o.key) +
		"&t=" + url.QueryEscape(title)
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
			ImdbRating string `json:"imdbRating"`
			Runtime    string `json:"Runtime"`
			Rated      string `json:"Rated"`
			Genre      string `json:"Poster"`
			Plot       string `json:"Plot"`
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

	results := make([]TitleSearchResult, 0, len(body.Search))
	for _, r := range body.Search {
		results = append(results, TitleSearchResult{
			ImdbRating: r.ImdbRating,
			Runtime:    r.Runtime,
			Rated:      r.Rated,
			Genre:      r.Genre,
			Plot:       r.Plot,
		})
	}
	return results, nil
}
