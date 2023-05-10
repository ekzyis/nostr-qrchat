package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Token struct {
	Name  string
	Value string
}

var (
	StaticServeDir                = "public"
	Tokens                        []Token
	NsecSessionKeyRegexp          = regexp.MustCompile(regexp.QuoteMeta(`""; /** NSEC SESSION KEY **/`))
	NpubSessionKeyRegexp          = regexp.MustCompile(regexp.QuoteMeta(`""; /** NPUB SESSION KEY **/`))
	NpubRecipientSessionKeyRegexp = regexp.MustCompile(regexp.QuoteMeta(`""; /** NPUB RECIPIENT KEY **/`))
)

func init() {
	file, err := os.Open("tokens")
	if err != nil {
		log.Fatal("error reading tokens: ", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		name := ""
		token := parts[0]
		if len(parts) > 1 {
			name = parts[0]
			token = parts[1]
		}
		Tokens = append(Tokens, Token{name, token})
	}
}

func addSession(indexFile []byte) (string, error) {
	keys, err := GenerateKeyPair()
	if err != nil {
		return "", fmt.Errorf("error generating session keys: %w", err)
	}
	privKey, pubKey := keys[0], keys[1]

	sessionFile := NsecSessionKeyRegexp.ReplaceAllString(string(indexFile), fmt.Sprintf(`"%s"`, privKey))
	sessionFile = NpubSessionKeyRegexp.ReplaceAllString(sessionFile, fmt.Sprintf(`"%s"`, pubKey))
	sessionFile = NpubRecipientSessionKeyRegexp.ReplaceAllString(sessionFile, fmt.Sprintf(`"%s"`, NostrPubKey))

	return sessionFile, nil
}

func loadIndexFile() ([]byte, error) {
	content, err := ioutil.ReadFile("public/index.html")
	if err != nil {
		return []byte{}, fmt.Errorf("error reading index.html: %w", err)
	}
	return content, nil
}

func handler(static http.Handler, session http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// try to serve static files
		_, err := http.Dir("public").Open(r.URL.Path)
		if err == nil {
			static.ServeHTTP(w, r)
			return
		}
		// else check if URL uses valid session token
		session.ServeHTTP(w, r)
	})
}

func checkToken(token string) bool {
	for _, token_ := range Tokens {
		if token_.Value == token {
			return true
		}
	}
	return false
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	indexFile, err := loadIndexFile()
	if err != nil {
		log.Println(err)
		http.Error(w, "error reading index.html", http.StatusInternalServerError)
		return
	}
	token := strings.TrimLeft(r.URL.Path, "/")
	valid := checkToken(token)
	if !valid {
		// Send index.html without session
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexFile)
		return
	}
	sessionFile, err := addSession(indexFile)
	if err != nil {
		log.Println(err)
		http.Error(w, "error generating session", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(sessionFile))
}

func main() {
	fs := http.FileServer(http.Dir(StaticServeDir))
	http.Handle("/", handler(fs, http.HandlerFunc(sessionHandler)))

	log.Println("Server started on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
