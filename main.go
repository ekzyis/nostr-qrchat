package main

import (
	"fmt"
	"log"
	"net/http"
)

func fileServerHandler(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("public"))
	fs.ServeHTTP(w, r)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Query().Get("msg")
	response := "Received message: " + message
	fmt.Fprint(w, response)
}

func main() {
	http.HandleFunc("/", fileServerHandler)
	http.HandleFunc("/chat", chatHandler)

	log.Println("Server started on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
