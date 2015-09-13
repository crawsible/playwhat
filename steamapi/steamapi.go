package steamapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type ResolveVanityURLResponse struct {
	Response struct {
		SteamID string `json:"steamid"`
		Success uint
	}
}

func ResolveVanityURL(steamName string) (*ResolveVanityURLResponse, error) {
	values := url.Values{}
	values.Add("vanityurl", url.QueryEscape(steamName))

	resolveVanityURLEndpoint := generateSteamAPIURL("ISteamUser/ResolveVanityURL/v0001", values, true)
	resolveVanityURLResponse := &ResolveVanityURLResponse{}
	err := decodeSteamAPIResponse(resolveVanityURLEndpoint, resolveVanityURLResponse)
	return resolveVanityURLResponse, err
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

type GetOwnedGamesResponse struct {
	Response struct {
		GameCount uint `json:"game_count"`
		Games     []Game
	}
}

func GetOwnedGames(steamID string) (*GetOwnedGamesResponse, error) {
	values := url.Values{}
	values.Add("steamid", url.QueryEscape(steamID))
	values.Add("include_appinfo", "1")

	getOwnedGamesEndpoint := generateSteamAPIURL("IPlayerService/GetOwnedGames/v0001", values, true)
	getOwnedGamesResponse := &GetOwnedGamesResponse{}
	err := decodeSteamAPIResponse(getOwnedGamesEndpoint, getOwnedGamesResponse)
	return getOwnedGamesResponse, err
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
