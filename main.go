package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "regexp"
    "time"

    "github.com/joho/godotenv"
    "github.com/PuerkitoBio/goquery"
)

type WishlistItem struct {
    Title string
    Price string
    URL   string
    ISBN  string
}

func scrapeWishlist(wishlistID string) ([]WishlistItem, error) {
    url := fmt.Sprintf("https://www.amazon.co.jp/hz/wishlist/ls/%s", wishlistID)
    fmt.Printf("Scraping URL: %s\n", url)

    client := &http.Client{
        Timeout: time.Second * 30,
    }

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request: %v", err)
    }
    defer resp.Body.Close()

    fmt.Printf("Response status: %s\n", resp.Status)

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error parsing HTML: %v", err)
    }

    var items []WishlistItem

    doc.Find("li[data-id]").Each(func(i int, s *goquery.Selection) {
        fmt.Printf("Processing item %d:\n", i+1)

        // 修正されたタイトルセレクタ
        title := s.Find("a[id^='itemName_']").Text()

        price := s.Find(".a-price .a-offscreen").First().Text()
        url, _ := s.Find("a[id^='itemName_']").Attr("href")
        
        // Extract ISBN
        isbn := ""
        re := regexp.MustCompile(`/dp/([A-Z0-9]{10})`)
        if matches := re.FindStringSubmatch(url); len(matches) > 1 {
            isbn = matches[1]
        }

        // Make URL absolute
        fullURL := "https://www.amazon.co.jp" + url

        fmt.Printf("Found item:\n")
        fmt.Printf("  Title: %s\n", strings.TrimSpace(title))
        fmt.Printf("  Price: %s\n", strings.TrimSpace(price))
        fmt.Printf("  URL: %s\n", fullURL)
        fmt.Printf("  ISBN: %s\n", isbn)
        fmt.Println()

        item := WishlistItem{
            Title: strings.TrimSpace(title),
            Price: strings.TrimSpace(price),
            URL:   fullURL,
            ISBN:  isbn,
        }
        items = append(items, item)
    })

    fmt.Printf("Total items found: %d\n", len(items))

    return items, nil
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    wishlistID := os.Getenv("AMAZON_WISHLIST_ID")
    if wishlistID == "" {
        log.Fatal("AMAZON_WISHLIST_ID is not set in .env file")
    }

    items, err := scrapeWishlist(wishlistID)
    if err != nil {
        log.Fatalf("Error scraping wishlist: %v", err)
    }

    fmt.Printf("Total items scraped: %d\n\n", len(items))

    for i, item := range items {
        fmt.Printf("Item %d:\n", i+1)
        fmt.Printf("Title: %s\nPrice: %s\nURL: %s\nISBN: %s\n\n", item.Title, item.Price, item.URL, item.ISBN)
    }
}
