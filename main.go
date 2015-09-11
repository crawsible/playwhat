package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

var templates = template.Must(template.ParseFiles("staticfiles/user_show.html"))

func userCreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello!")
	fmt.Println("The received method is... ", r.Method)
	switch r.Method {
	case "POST":
		r.ParseForm()

		u := User{SteamName: r.PostFormValue("steamname")}
		u.fetchSteamID()
		fmt.Println("Your SteamID is...", u.SteamID)

		u.fetchOwnedGames()
		fmt.Println("These are the games you own... ", u.Games)
		fmt.Printf("You own %d games\n", len(u.Games))

		templates.ExecuteTemplate(w, "user_show.html", u)
	default:
		http.NotFound(w, r)
	}
}

type User struct {
	SteamName string
	SteamID   string
	Games     []Game
}

func (u *User) fetchSteamID() (err error) {
	u.SteamID, err = resolveVanityURL(u.SteamName)
	return
}

func (u *User) fetchOwnedGames() (err error) {
	u.Games, err = getOwnedGames(u.SteamID)
	return
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

type Game struct {
	Name     string
	AppID    int
	Playtime int `json:"playtime_forever"`

	LogoFilename string `json:"img_logo_url"`
}

func (g *Game) LogoURL() string {
	if g.AppID == 0 || g.LogoFilename == "" {
		return "http://digilite.ca/wp-content/uploads/2013/07/squarespace-184x69.jpg"
	}

	return fmt.Sprintf(
		"http://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg",
		g.AppID,
		g.LogoFilename,
	)
}

type GetOwnedGamesResponse struct {
	Response struct {
		Games []Game
	}
}

func getOwnedGames(steamID string) ([]Game, error) {
	values := url.Values{}
	values.Add("steamid", url.QueryEscape(steamID))
	values.Add("include_appinfo", "1")

	getOwnedGamesEndpoint := generateSteamAPIURL("IPlayerService/GetOwnedGames/v0001", values, true)

	resp, err := http.Get(getOwnedGamesEndpoint.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))

	structuredResponse := &GetOwnedGamesResponse{}
	err = json.Unmarshal(body, structuredResponse)
	if err != nil {
		return nil, err
	}

	return structuredResponse.Response.Games, nil
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
