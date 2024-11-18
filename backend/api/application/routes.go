package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/DanVerh/university-swe/backend/api/handlers"
)

// Create router with confgiured routes
func loadRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/products", loadProductsRoutes)
	router.Route("/customers", loadCustomersRoutes)
	//router.Route("/orders", loadOrdersRoutes)

	return router
}

// Define all routes with HTTP methods
func loadProductsRoutes(router chi.Router) {
	productsHandler := &handlers.ProductsHandler{}
	router.Post("/", productsHandler.Create)
	router.Get("/", productsHandler.List)
	router.Get("/{id}", productsHandler.GetByID)
	router.Put("/{id}", productsHandler.UpdateByID)
	router.Delete("/{id}", productsHandler.DeleteByID)
}

func loadCustomersRoutes(router chi.Router) {
	customersHandler := &handlers.CustomersHandler{}
	router.Post("/", customersHandler.Create)
	router.Get("/", customersHandler.List)
	router.Get("/{id}", customersHandler.GetByID)
	router.Put("/{id}", customersHandler.UpdateByID)
	router.Delete("/{id}", customersHandler.DeleteByID)
}

/*func loadOrdersRoutes(router chi.Router) {
	ordersHandler := &handler.OrdersHandler{}
	router.Post("/", ordersHandler.Create)
	router.Get("/", ordersHandler.List)
	router.Get("/{id}", ordersHandler.GetByID)
	router.Put("/{id}", ordersHandler.UpdateByID)
	router.Delete("/{id}", ordersHandler.DeleteByID)
}*/