package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func userCreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello!")
	fmt.Println("The received method is... ", r.Method)
	switch r.Method {
	case "POST":
		r.ParseForm()

		steamID, _ := resolveVanityURL(r.PostFormValue("steamname"))
		fmt.Println("Your SteamID is...", steamID)

		http.Redirect(w, r, "/", http.StatusFound)
	default:
		http.NotFound(w, r)
	}
}

type ResolveVanityURLResponse struct {
	Response struct {
		SteamID string `json:"steamid"`
	}
}

func resolveVanityURL(steamName string) (string, error) {
	values := url.Values{}
	values.Add("vanityurl", url.QueryEscape(steamName))

	resolveVanityURLEndpoint := generateSteamAPIURL("ISteamUser/ResolveVanityURL/v0001", values, true)

	resp, err := http.Get(resolveVanityURLEndpoint.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println(string(body))

	structuredResponse := &ResolveVanityURLResponse{}
	err = json.Unmarshal(body, structuredResponse)
	if err != nil {
		return "", err
	}

	return structuredResponse.Response.SteamID, nil
}

func generateSteamAPIURL(apiPath string, values url.Values, withKey bool) *url.URL {
	generatedURL := &url.URL{Scheme: "http", Host: "api.steampowered.com", Path: apiPath}

	if withKey {
		values.Add("key", os.Getenv("STEAM_API_KEY"))
	}
	generatedURL.RawQuery = values.Encode()

	fmt.Println("the URL is...", generatedURL.String())
	return generatedURL
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./staticfiles")))
	http.HandleFunc("/user/create", userCreateHandler)

	http.ListenAndServe(":8080", nil)
}
