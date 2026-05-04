package main

import (
	"log"
	"net/http"

	"myweb/database"
	"myweb/routes"
)

func main() {
	database.Connect()
	routes.SetupRoutes()

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
