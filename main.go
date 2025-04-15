package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

type URLMap struct {
	urls  map[string]string
	mutex sync.RWMutex
}

func main() {
	godotenv.Load(".env")

	urlMap := &URLMap{
		urls: make(map[string]string),
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.HandleFunc("/health", handleHealthCheck)
	router.HandleFunc("POST /shorten", urlMap.handleEncodeRequest)
	router.HandleFunc("GET /g/{shortUrl}", urlMap.redirectToOriginaUrl)

	fmt.Printf("Server running on PORT:%s\n", os.Getenv("PORT"))

	http.ListenAndServe(":"+os.Getenv("PORT"), router)
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{
		Message: "Health Check",
	})
}

func (urlMap *URLMap) redirectToOriginaUrl(w http.ResponseWriter, r *http.Request) {
	shortUrl := chi.URLParam(r, "shortUrl")
	if shortUrl == "" {
		w.WriteHeader(400)
		w.Write([]byte("Invalid URL"))
		return
	}

	originalUrl, ok := urlMap.urls[shortUrl]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalUrl, http.StatusFound)
}

func (urlMap *URLMap) handleEncodeRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	shortenedUrl := EncodeUrl(req.URL)

	urlMap.mutex.Lock()
	urlMap.urls[shortenedUrl] = req.URL
	urlMap.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(struct {
		Message      string `json:"message"`
		ShortenedUrl string `json:"shortened_url"`
	}{
		Message:      "Ok",
		ShortenedUrl: fmt.Sprintf("%s/go/%s", os.Getenv("APP_URL"), shortenedUrl),
	})
}
