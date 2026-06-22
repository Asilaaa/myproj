package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"myproj/internal/ai"
	"myproj/internal/auth"
	"myproj/internal/config"
	"myproj/internal/database"
	"myproj/internal/httpx"
	"myproj/internal/images"
	"myproj/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	minioClient, err := storage.NewClient(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOUseSSL)
	if err != nil {
		log.Fatal(err)
	}

	storageService := storage.NewService(minioClient, cfg.MinIOBucket)
	aiService := ai.NewService(cfg.OpenAIAPIKey)
	imageRepository := images.NewRepository(db)
	imageService := images.NewService(imageRepository, storageService, aiService, cfg.MaxUploadSizeBytes)
	imageHandler := images.NewHandler(imageService)

	oryClient := auth.NewClient(cfg.OryURL)
	authService := auth.NewService(oryClient)
	authMiddleware := auth.NewMiddleware(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	mux.Handle("GET /api/me", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetUserIDFromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "missing user in context")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, map[string]string{"user_id": userID})
	})))
	mux.Handle("GET /api/images", authMiddleware.RequireAuth(http.HandlerFunc(imageHandler.GetImages)))
	mux.Handle("POST /api/images/upload", authMiddleware.RequireAuth(http.HandlerFunc(imageHandler.UploadImage)))
	mux.Handle("POST /api/images/import", authMiddleware.RequireAuth(http.HandlerFunc(imageHandler.ImportImage)))
	mux.Handle("GET /api/images/{id}", authMiddleware.RequireAuth(http.HandlerFunc(imageHandler.GetImage)))
	mux.Handle("DELETE /api/images/{id}", authMiddleware.RequireAuth(http.HandlerFunc(imageHandler.DeleteImage)))
	mux.Handle("GET /api/objects", authMiddleware.RequireAuth(http.HandlerFunc(imageHandler.ListBucketObjects)))

	handler := httpx.CORS(mux, cfg.FrontendOrigin)
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	log.Printf("API server running on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
