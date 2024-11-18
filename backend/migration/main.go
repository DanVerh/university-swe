package main

import (
	"github.com/DanVerh/university-swe/backend/migration/application"
)

func main() {
	app := application.New()
	app.Start()
}