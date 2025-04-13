package main

import (
	"fmt"
	"net/http"
	"os"
)

var SERVER_HOST = os.Getenv("SERVER_HOST")
var SERVER_PORT = os.Getenv("SERVER_PORT")

func base(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%v:%v", SERVER_HOST, SERVER_PORT)
}

func main() {

	if SERVER_HOST == "" {
		fmt.Println("SERVER_HOST is not set")
	}

	if SERVER_PORT == "" {
		fmt.Println("SERVER_PORT is not set")
	}

	http.HandleFunc("/", base)

	fmt.Printf("Server listening on %v:%v...\n", SERVER_HOST, SERVER_PORT)
	http.ListenAndServe(":"+SERVER_PORT, nil)
}
