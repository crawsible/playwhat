package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/crawsible/playwhat/steamapi"
)

var templates = template.Must(template.ParseFiles("view/user_show.html"))

type User struct {
	SteamName string
	SteamID   string
	Games     Games
}

type Games []steamapi.Game

func (gs Games) Len() int {
	return len(gs)
}

func (gs Games) Less(i, j int) bool {
	return gs[i].Playtime < gs[j].Playtime
}

func (gs Games) Swap(i, j int) {
	gs[i], gs[j] = gs[j], gs[i]
}

func userCreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello!")
	fmt.Println("The received method is... ", r.Method)
	switch r.Method {
	case "POST":
		r.ParseForm()

		u := User{SteamName: r.PostFormValue("steamname")}

		resolveVanityURLResponse, err := steamapi.ResolveVanityURL(u.SteamName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u.SteamID = resolveVanityURLResponse.Response.SteamID
		fmt.Println("Your SteamID is...", u.SteamID)

		var getOwnedGamesResponse *steamapi.GetOwnedGamesResponse
		getOwnedGamesResponse, err = steamapi.GetOwnedGames(u.SteamID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u.Games = getOwnedGamesResponse.Response.Games
		sort.Sort(sort.Reverse(u.Games))
		fmt.Println("These are the games you own... ", u.Games)

		templates.ExecuteTemplate(w, "user_show.html", u)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./view")))
	http.HandleFunc("/user/create", userCreateHandler)

	http.ListenAndServe(":8080", nil)
}
