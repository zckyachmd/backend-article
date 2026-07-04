package main

import (
	"log"
	"net/http"

	"backend-article/internal/article"
	"backend-article/internal/config"
	"backend-article/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := article.NewRepository(db)
	service := article.NewService(repo)
	handler := article.NewHandler(service)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	article.RegisterRoutes(router, handler)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	log.Printf("server running on :%s", cfg.AppPort)
	log.Fatal(server.ListenAndServe())
}
