package main

import (
	"fmt"
	"net/http"
)

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received callback request")
	fmt.Fprintln(w, "Callback received successfully!")
}

func main_() {
	http.HandleFunc("/callback", callbackHandler)
	fmt.Println("Local server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
