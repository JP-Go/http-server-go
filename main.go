package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/JP-Go/http-server-go/internal/api"
	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		fmt.Println("Misconfigured environment. Missing variable DB_URL")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Println("Could not connect to database. Exiting")
		os.Exit(1)
	}

	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	chirpyApi := api.NewApi(database.New(db))
	chirpyApi.RegisterEndpoints(fileServer, mux)
	chirpyApi.Serve(mux)
}
