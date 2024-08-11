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

    wishlistID := os.Getenv("AMAZON_WISHLIST_ID")
    if wishlistID == "" {
        log.Fatal("AMAZON_WISHLIST_ID is not set in .env file")
    }

    log.Printf("Starting scraping for wishlist ID: %s\n", wishlistID)

    items, err := ScrapeWishlist(wishlistID)
    if err != nil {
        log.Fatalf("Error scraping wishlist: %v", err)
    }

    log.Printf("Total unique items scraped: %d\n\n", len(items))

    err = SaveToDatabase(items)
    if err != nil {
        log.Fatalf("Error saving to database: %v", err)
    }

    log.Println("Data successfully saved to database")

    for i, item := range items {
        fmt.Printf("Item %d:\n", i+1)
        fmt.Printf("Title: %s\nPrice: %s\nURL: %s\nISBN: %s\n\n", item.Title, item.Price, item.URL, item.ISBN)
    }
}
