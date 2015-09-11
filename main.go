package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"sort"
)

var templates = template.Must(template.ParseFiles("view/user_show.html"))

func userCreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello!")
	fmt.Println("The received method is... ", r.Method)
	switch r.Method {
	case "POST":
		r.ParseForm()

		u := User{SteamName: r.PostFormValue("steamname")}
		if err := u.FetchSteamID(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("Your SteamID is...", u.SteamID)

		if err := u.FetchOwnedGames(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sort.Sort(sort.Reverse(u.Games))
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
	Games     Games
}

func (u *User) FetchSteamID() (err error) {
	u.SteamID, err = resolveVanityURL(u.SteamName)
	return
}

func (u *User) FetchOwnedGames() (err error) {
	u.Games, err = getOwnedGames(u.SteamID)
	return
}

type ResolveVanityURLResponse struct {
	Response struct {
		SteamID string `json:"steamid"`
		Success uint
	}
}

func resolveVanityURL(steamName string) (string, error) {
	values := url.Values{}
	values.Add("vanityurl", url.QueryEscape(steamName))

	resolveVanityURLEndpoint := generateSteamAPIURL("ISteamUser/ResolveVanityURL/v0001", values, true)
	vanityURLResponse := &ResolveVanityURLResponse{}
	if err := decodeSteamAPIResponse(resolveVanityURLEndpoint, vanityURLResponse); err != nil {
		return "", err
	}

	return vanityURLResponse.Response.SteamID, nil
}

type Game struct {
	Name                     string
	AppID                    uint
	Playtime                 uint   `json:"playtime_forever"`
	LogoImageFilename        string `json:"img_logo_url"`
	IconImageFilename        string `json:"img_icon_url"`
	HasCommunityVisibleStats bool   `json:"has_community_visible_stats"`
}

func (g *Game) LogoURL() string {
	if g.AppID == 0 || g.LogoImageFilename == "" {
		return "http://digilite.ca/wp-content/uploads/2013/07/squarespace-184x69.jpg"
	}

	return fmt.Sprintf(
		"http://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg",
		g.AppID,
		g.LogoImageFilename,
	)
}

type Games []Game

func (gs Games) Len() int {
	return len(gs)
}

func (gs Games) Less(i, j int) bool {
	return gs[i].Playtime < gs[j].Playtime
}

func (gs Games) Swap(i, j int) {
	gs[i], gs[j] = gs[j], gs[i]
}

type GetOwnedGamesResponse struct {
	Response struct {
		GameCount uint `json:"game_count"`
		Games
	}
}

func getOwnedGames(steamID string) (Games, error) {
	values := url.Values{}
	values.Add("steamid", url.QueryEscape(steamID))
	values.Add("include_appinfo", "1")

	getOwnedGamesEndpoint := generateSteamAPIURL("IPlayerService/GetOwnedGames/v0001", values, true)
	ownedGamesResponse := &GetOwnedGamesResponse{}
	if err := decodeSteamAPIResponse(getOwnedGamesEndpoint, ownedGamesResponse); err != nil {
		return nil, err
	}

	return ownedGamesResponse.Response.Games, nil
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

func decodeSteamAPIResponse(apiURL *url.URL, data interface{}) error {
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

func main() {
	http.Handle("/", http.FileServer(http.Dir("./view")))
	http.HandleFunc("/user/create", userCreateHandler)

	http.ListenAndServe(":8080", nil)
}
