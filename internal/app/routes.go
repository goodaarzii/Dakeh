package app

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/goodaarzii/dakeh/internal/json"
)

func (app *Application) Mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		json.Write(w, http.StatusOK, []byte("all good"))
	})

	//products := product.NewHandler(product.NewService(repo.New(app.Db)))
	//r.Get("/products", products.ListProducts)
	//r.Get("/product/{id}", products.GetProductByID)

	return r
}

// Run starts the HTTP server
func (app *Application) Run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      h,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on %s", app.Config.Addr)
	return srv.ListenAndServe()
}
