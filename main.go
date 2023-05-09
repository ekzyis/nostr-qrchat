package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var (
	NsecSessionKeyRegexp          = regexp.MustCompile(regexp.QuoteMeta(`"" /** NSEC SESSION KEY **/`))
	NpubSessionKeyRegexp          = regexp.MustCompile(regexp.QuoteMeta(`"" /** NPUB SESSION KEY **/`))
	NpubRecipientSessionKeyRegexp = regexp.MustCompile(regexp.QuoteMeta(`"" /** NPUB RECIPIENT KEY **/`))
)

func generateSession() (string, error) {
	content, err := ioutil.ReadFile("public/index.html")
	if err != nil {
		return "", fmt.Errorf("error reading index.html: %w", err)
	}

	keys, err := GenerateKeyPair()
	if err != nil {
		return "", fmt.Errorf("error generating session keys: %w", err)
	}
	privKey, pubKey := keys[0], keys[1]

	modifiedContent := NsecSessionKeyRegexp.ReplaceAllString(string(content), fmt.Sprintf(`"%s"`, privKey))
	modifiedContent = NpubSessionKeyRegexp.ReplaceAllString(modifiedContent, fmt.Sprintf(`"%s"`, pubKey))
	modifiedContent = NpubRecipientSessionKeyRegexp.ReplaceAllString(modifiedContent, fmt.Sprintf(`"%s"`, NostrPubKey))

	return modifiedContent, nil
}

func fileHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// intercept the request for the index.html file
			// to generate a new session
			indexFile, err := generateSession()
			if err != nil {
				log.Println(err)
				http.Error(w, "Failed to generate session", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(indexFile))
			return
		}

		// all other requests are passed to the underlying file server
		h.ServeHTTP(w, r)
	})
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Query().Get("msg")
	response := "Received message: " + message
	fmt.Fprint(w, response)
}

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fileHandler(fs))
	http.HandleFunc("/chat", chatHandler)

	log.Println("Server started on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
