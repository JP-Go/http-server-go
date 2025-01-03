package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/JP-Go/http-server-go/internal/api"
	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func MustLoadEnv(envVariable string) string {
	variable := os.Getenv(envVariable)
	if variable == "" {
		log.Fatalf("Misconfigured environment. Missing variable %s", envVariable)
	}
	return variable
}
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func main() {
	godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("Empty JWT secret. Using the string 'secret'. THIS MUST NOT BE USED IN PRODUCTION")
		jwtSecret = "secret"
	}
	portNumber := 8080
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("Empty PORT variable. Using 8080")
	} else {
		possiblePort, err := strconv.Atoi(port)
		if err != nil {
			fmt.Println("Invalid PORT number. Using 8080")
		} else {
			portNumber = possiblePort
		}
	}

	dbUrl := MustLoadEnv("DB_URL")
	db := Must(sql.Open("postgres", dbUrl))

	polkaApiKey := MustLoadEnv("POLKA_KEY")

	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	apiConfig := api.ApiConfig{
		DB:          database.New(db),
		PolkaApiKey: polkaApiKey,
		JwtSecret:   jwtSecret,
	}
	chirpyApi := api.NewApi(&apiConfig)
	chirpyApi.RegisterEndpoints(fileServer, mux)
	chirpyApi.Serve(mux, portNumber)
}
