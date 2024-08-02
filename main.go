package main

import (
    "fmt"
    "log"
    "os"
    "github.com/joho/godotenv"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    whishlistID := os.Getenv("AMAZON_WHISHLIST_ID")
}
