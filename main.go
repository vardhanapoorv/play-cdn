package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Toy CDN

// - Fetch requested static file from filesystem that is cache for our CDN
// - If file is not found, fetch from origin server and cache it
// - Serve the file to the client

// To achieve the above -
// - CDN needs to store the origin server URL
// - CDN needs to store the static files on first response from origin server

var originCDNMappings = map[string]string{
	"localhost:8081": "https://www.apoorvvardhan.dev",
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", getStaticFile)

	log.Print("Starting server on port 8081")

	http.ListenAndServe(":8081", mux)
}

func getStaticFile(w http.ResponseWriter, r *http.Request) {

	println("Request received for", r.URL.Path)
	originServer := originCDNMappings[r.Host]
	url := []byte(originServer + r.URL.Path)

	url_hash := md5.Sum(url)
	stringUrlHash := hex.EncodeToString(url_hash[:])

	filePath := ".cache/" + stringUrlHash
	content, _ := os.ReadFile(filePath)
	if content != nil {
		println("File found in cache")
		http.ServeFile(w, r, filePath)
		return
	} else {
		// Fetch file from origin server
		println("File not found in cache")
		resp, err := http.Get(originServer + r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		content, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body", err)
		}
		resp.Body.Close()
		err = os.WriteFile(filePath, content, 0644)
		if err != nil {
			fmt.Println("Error writing file to cache", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
