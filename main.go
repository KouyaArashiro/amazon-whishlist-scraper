package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "github.com/joho/godotenv"
    "github.com/PuerkitoBio/goquery"
)

type WishlistItem struct {
    Title string
    Price string
    URL   string
}

func scrapeWishlist(wishlistID string) ([]WishlistItem, error) {
    url := fmt.Sprintf("https://www.amazon.co.jp/zn/wishlist/ls/%s", wishlistID)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, err
    }

    var items []WishlistItem
    //Implement scraping logic
    doc.Find("li[data-id]").Each(func(i int, s *goquery.Selection) {
        title := s.Find("h2").Text()
        price := s.Find(".a-price-whole").First().Text()
        url, _ := s.Find("a").Attr("href")
        
        item := WishlistItem{
            Title: title,
            Price: price,
            URL:   url,
        }
        items = append(items, item)
    })

    return items, nil
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    wishlistID := os.Getenv("AMAZON_WISHLIST_ID")

    items, err := scrapeWishlist(wishlistID)
    if err != nil {
        log.Fatalf("Error scraping whislist: %v", err)
    }

    for _, item := range items {
        fmt.Println("Title: %s, Price: %s, url: %s\n", item.Title, item.Price, item.URL)
    }
}
