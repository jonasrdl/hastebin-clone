package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jonasrdl/hastebin-clone/handlers"
	"os"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	apiKey := os.Getenv("API_KEY")

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dbConnectionString)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pastes (
			ID VARCHAR(36) PRIMARY KEY,
			Content TEXT,
			CreatedAt DATETIME,
			Password VARCHAR(16)
		);
	`)
	if err != nil {
		fmt.Printf("Error creating table: %v\n", err)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	pasteHandler := handlers.NewPasteHandler(db, apiKey)

	api := r.Group("/")
	{
		api.POST("/", pasteHandler.CreatePaste)
		api.GET("/:id", pasteHandler.GetPaste)
	}

	err = r.Run(":" + port)
	if err != nil {
		fmt.Printf("error running api: %v", err)
		return
	}
}
