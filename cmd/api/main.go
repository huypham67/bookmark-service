package main

import (
	"github.com/huypham67/bookmark-service/internal/bootstrap"

	// Docs package is required to automatically register Swagger documentation via its init() function.
	_ "github.com/huypham67/bookmark-service/docs"
)

// @title Bookmark Service API
// @version 1.1
// @description This is the API documentation for the Bookmark Service

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
