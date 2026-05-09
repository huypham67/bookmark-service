package main

import (
	"github.com/huypham67/bookmark-management/internal/bootstrap"

	_ "github.com/huypham67/bookmark-management/docs"
)

// @title Bookmark Management API
// @version 1.0
// @description This is the API documentation for the Bookmark Management service.
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		panic(err)
	}

	if err := app.Run(); err != nil {
		panic(err)
	}
}
