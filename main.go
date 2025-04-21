package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type myRedis struct {
	myRedis *redis.Client
}

func main() {
	godotenv.Load(".env")

	red := &myRedis{
		myRedis: redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_URL"),
			Password: "",
			DB:       0,
		}),
	}

	val := red.myRedis.Ping(ctx)
	fmt.Println(val)

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.HandleFunc("/health", handleHealthCheck)
	router.Post("/shorten", red.handleShortenRequest)
	router.Get("/g/{shortUrl}", red.redirectToOriginalUrl)

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

func (red *myRedis) redirectToOriginalUrl(w http.ResponseWriter, r *http.Request) {
	shortUrl := chi.URLParam(r, "shortUrl")
	if shortUrl == "" {
		w.WriteHeader(400)
		w.Write([]byte("Invalid URL"))
		return
	}

	originalUrl, err := red.myRedis.Get(ctx, shortUrl).Result()

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.Redirect(w, r, originalUrl, http.StatusFound)
}

func (red *myRedis) handleShortenRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	shortenedUrl := EncodeUrl(req.URL)

	err := red.myRedis.Set(ctx, shortenedUrl, req.URL, 20*time.Second).Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(struct {
		Message      string `json:"message"`
		ShortenedUrl string `json:"shortened_url"`
	}{
		Message:      "Ok",
		ShortenedUrl: fmt.Sprintf("%s/g/%s", os.Getenv("APP_URL"), shortenedUrl),
	})
}
