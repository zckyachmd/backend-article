package article

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler Handler) {
	r.Route("/article", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/{limit}/{offset}", handler.List)
		r.Get("/{id}", handler.GetByID)
		r.Put("/{id}", handler.Update)
		r.Patch("/{id}", handler.Patch)
		r.Delete("/{id}", handler.Delete)
	})
}
