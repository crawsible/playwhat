package gbapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type Game struct {
	Name string
	Deck string
}

type SearchResponse struct {
	Games []Game `json:"results"`
}

func Search(query string) (*SearchResponse, error) {
	values := url.Values{}
	values.Add("query", url.QueryEscape(query))
	values.Add("resources", "game")
	values.Add("field_list", "name")
	values.Add("field_list", "deck")

	searchResponseEndpoint := generateURL("api/search", values)
	searchResponse := &SearchResponse{}
	err := decodeResponse(searchResponseEndpoint, searchResponse)
	return searchResponse, err
}

func generateURL(apiPath string, values url.Values) *url.URL {
	generatedURL := &url.URL{Scheme: "http", Host: "www.giantbomb.com", Path: apiPath}

	values.Add("api_key", os.Getenv("GIANTBOMB_API_KEY"))
	values.Add("format", "json")
	generatedURL.RawQuery = values.Encode()

	fmt.Println("the URL is...", generatedURL.String())
	return generatedURL
}

func decodeResponse(apiURL *url.URL, data interface{}) error {
	r, err := http.Get(apiURL.String())
	if err != nil {
		return err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		return err
	}

	return nil
}
