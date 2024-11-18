package application

import (
	"log"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
)

// Create application struct (class) with required for migration fields
type App struct {
	file string
	dbUri string
}

// Construct for the App object
func New() *App {
	app := &App{
		file: "file://migrations",
		dbUri: "mongodb://localhost:27017/sales",
	}

	return app
}

// Start the app by running the migration
func (app *App) Start() error {
	m, err := migrate.New(
    	app.file,
        app.dbUri,
    )
    if err != nil {
        log.Fatalf("Failed to create migrate instance: %v", err)
    }

	err = m.Up()
    if err != nil && err != migrate.ErrNoChange {
        log.Fatalf("Failed to run up migrations: %v", err)
    }
    fmt.Println("Migrations applied successfully")

	return nil
}