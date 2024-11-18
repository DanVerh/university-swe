package application

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// Define port constant value
const port int = 8080

// Define App struct (class)
type App struct {
	router http.Handler
}

// Define constructor for creating object of App class
// Pointer, because we need to modify object fields
func New() *App {
	app := &App{
		router: loadRoutes(),
	}

	return app
}

// Method for starting the app server
func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port), // convert port to ASCII
		Handler: app.router,
	}
	
	fmt.Printf("Application started on localhost:%d\n", port)

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
