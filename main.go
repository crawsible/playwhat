package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./staticfiles")))
	http.ListenAndServe(":8080", nil)
}
