package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Starting file server.")
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8080", nil)
}
