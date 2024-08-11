package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/chromedp/chromedp"
    "github.com/joho/godotenv"
    _ "github.com/go-sql-driver/mysql"
)

//type WishlistItem struct {
//    Title string
//    Price string
//    URL   string
//    ISBN  string
//}

func scrapeWishlist(wishlistID string) ([]WishlistItem, error) {
log.Println("Starting scrapeWishlist function")

    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )

    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    defer cancel()

    ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
    defer cancel()

    ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    var items []WishlistItem
    uniqueItems := make(map[string]bool) // ISBNをキーとして使用

    err := chromedp.Run(ctx,
        chromedp.Navigate(fmt.Sprintf("https://www.amazon.co.jp/hz/wishlist/ls/%s", wishlistID)),
        chromedp.ActionFunc(func(ctx context.Context) error {
            log.Println("Page loaded, waiting for #g-items")
            return nil
        }),
        chromedp.WaitVisible(`#g-items`, chromedp.ByID),
        chromedp.ActionFunc(func(ctx context.Context) error {
            log.Println("#g-items found, starting to scroll and scrape")
            return scrollAndScrape(ctx, &items, uniqueItems)
        }),
    )

    if err != nil {
        return nil, fmt.Errorf("error in chromedp.Run: %v", err)
    }

    log.Printf("Scraping completed. Total unique items found: %d\n", len(items))
    return items, nil
}

func scrollAndScrape(ctx context.Context, items *[]WishlistItem, uniqueItems map[string]bool) error {
    var lastHeight int64
    for {
        var currentHeight int64
        if err := chromedp.Evaluate(`document.documentElement.scrollHeight`, &currentHeight).Do(ctx); err != nil {
            return fmt.Errorf("error getting page height: %v", err)
        }

        log.Printf("Current height: %d, Last height: %d\n", currentHeight, lastHeight)

        if currentHeight == lastHeight {
            log.Println("Reached end of page")
            break
        }

        if err := chromedp.Evaluate(`window.scrollTo(0, document.documentElement.scrollHeight)`, nil).Do(ctx); err != nil {
            return fmt.Errorf("error scrolling: %v", err)
        }

        time.Sleep(2 * time.Second)

        var newItems []WishlistItem
        err := chromedp.Evaluate(`
            Array.from(document.querySelectorAll('li[data-id]')).map(item => {
                let title = item.querySelector('a[id^="itemName_"]');
                let price = item.querySelector('.a-price .a-offscreen');
                let url = title ? title.href : '';
                let isbn = url.match(/\/dp\/(\d{10}|\d{13})/) ? url.match(/\/dp\/(\d{10}|\d{13})/)[1] : '';
                return {
                    Title: title ? title.textContent.trim() : '',
                    Price: price ? price.textContent.trim() : '',
                    URL: url,
                    ISBN: isbn
                };
            })
        `, &newItems).Do(ctx)

        if err != nil {
            return fmt.Errorf("error evaluating JavaScript: %v", err)
        }

        for _, item := range newItems {
            if item.ISBN != "" && !uniqueItems[item.ISBN] {
                uniqueItems[item.ISBN] = true
                *items = append(*items, item)
                log.Printf("Added new item: %s (ISBN: %s)\n", item.Title, item.ISBN)
            } else if item.ISBN == "" {
                log.Printf("Skipped non-book item: %s\n", item.Title)
            } else {
                log.Printf("Skipped duplicate item: %s (ISBN: %s)\n", item.Title, item.ISBN)
            }
        }

        lastHeight = currentHeight
    }

    return nil
}

func saveToDatabase(items []WishlistItem) error {
    // データベース接続情報を環境変数から取得
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASS")
    dbName := os.Getenv("DB_NAME")
    dbHost := os.Getenv("DB_HOST")

    // データベース接続
    db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbName))
    if err != nil {
        return fmt.Errorf("error connecting to database: %v", err)
    }
    defer db.Close()

    // 接続テスト
    err = db.Ping()
    if err != nil {
        return fmt.Errorf("error pinging database: %v", err)
    }

    // トランザクション開始
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %v", err)
    }
    defer tx.Rollback() // トランザクションが正常にコミットされなかった場合はロールバック

    // 既存のタイトルを取得
    existingTitles, err := getExistingTitles(tx)
    if err != nil {
        return fmt.Errorf("error getting existing titles: %v", err)
    }

    // プリペアードステートメントの作成
    stmt, err := tx.Prepare("INSERT INTO wishlist_items(title, price, url, isbn) VALUES(?, ?, ?, ?)")
    if err != nil {
        return fmt.Errorf("error preparing statement: %v", err)
    }
    defer stmt.Close()

    // データの挿入
    insertedCount := 0
    skippedCount := 0
    for _, item := range items {
        if _, exists := existingTitles[item.Title]; exists {
            log.Printf("Skipping duplicate title: %s", item.Title)
            skippedCount++
            continue
        }

        _, err = stmt.Exec(item.Title, item.Price, item.URL, item.ISBN)
        if err != nil {
            return fmt.Errorf("error inserting item %s: %v", item.Title, err)
        }
        insertedCount++
    }

    // トランザクションのコミット
    err = tx.Commit()
    if err != nil {
        return fmt.Errorf("error committing transaction: %v", err)
    }

    log.Printf("Inserted %d new items, skipped %d duplicates", insertedCount, skippedCount)
    return nil
}

func getExistingTitles(tx *sql.Tx) (map[string]bool, error) {
    rows, err := tx.Query("SELECT title FROM wishlist_items")
    if err != nil {
        return nil, fmt.Errorf("error querying existing titles: %v", err)
    }
    defer rows.Close()

    existingTitles := make(map[string]bool)
    for rows.Next() {
        var title string
        if err := rows.Scan(&title); err != nil {
            return nil, fmt.Errorf("error scanning title: %v", err)
        }
        existingTitles[title] = true
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %v", err)
    }

    return existingTitles, nil
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

    log.Printf("Starting scraping for wishlist ID: %s\n", wishlistID)

    items, err := scrapeWishlist(wishlistID)
    if err != nil {
        log.Fatalf("Error scraping wishlist: %v", err)
    }

    log.Printf("Total unique items scraped: %d\n\n", len(items))

    err = saveToDatabase(items)
    if err != nil {
        log.Fatalf("Error saving to database: %v", err)
    }

    log.Println("Data successfully saved to database")

    for i, item := range items {
        fmt.Printf("Item %d:\n", i+1)
        fmt.Printf("Title: %s\nPrice: %s\nURL: %s\nISBN: %s\n\n", item.Title, item.Price, item.URL, item.ISBN)
    }
}
