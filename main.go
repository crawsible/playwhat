package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/crawsible/playwhat/steamapi"
	_ "github.com/mattn/go-sqlite3"
)

var userTemplates = template.Must(template.ParseFiles("view/user/show.html"))
var gameTemplates = template.Must(template.ParseFiles("view/game/show.html"))

type User struct {
	SteamName string
	SteamID   string
	Games     Games
}

func (u *User) getSteamID() error {
	db, err := sql.Open("sqlite3", "./db/dev.sqlite3.db")
	if err != nil {
		return err
	}

	row := db.QueryRow("SELECT id, steam_id FROM users WHERE steam_name = ?", u.SteamName)
	var id uint
	if err := row.Scan(&id, &u.SteamID); err != nil || u.SteamID == "" {
		fmt.Printf("err: %s", err.Error())
		fmt.Println("User ID not cached. Querying...")
	} else {
		fmt.Println("User SteamID found in cache!")
		return nil
	}

	resolveVanityURLResponse, err := steamapi.ResolveVanityURL(u.SteamName)
	if err != nil {
		return err
	}

	u.SteamID = resolveVanityURLResponse.Response.SteamID
	fmt.Println("Your SteamID is...", u.SteamID)

	var stmt *sql.Stmt
	if id != 0 {
		stmt, err = db.Prepare("UPDATE users SET steam_id = ? WHERE id = ?")
		if err != nil {
			fmt.Printf("Could not add steam_id to existing `users` record.")
			return nil
		}

		_, err := stmt.Exec(u.SteamID, id)
		return err
	} else {
		stmt, err = db.Prepare("INSERT INTO users(steam_name, steam_id) values(?, ?)")
		_, err := stmt.Exec(u.SteamName, u.SteamID)
		return err
	}
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

		if err := u.getSteamID(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var getOwnedGamesResponse *steamapi.GetOwnedGamesResponse
		getOwnedGamesResponse, err := steamapi.GetOwnedGames(u.SteamID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u.Games = getOwnedGamesResponse.Response.Games
		sort.Sort(sort.Reverse(u.Games))
		fmt.Println("These are the games you own... ", u.Games)

		userTemplates.ExecuteTemplate(w, "show.html", u)
	default:
		http.NotFound(w, r)
	}
}

func gameShowHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Game!")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./view")))
	http.HandleFunc("/user/create", userCreateHandler)
	http.HandleFunc("/user/game", gameShowHandler)

	http.ListenAndServe(":8080", nil)
}
