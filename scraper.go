package main

import (
    "fmt"
    "log"
    "context"
    "time"
    
    "github.com/chromedp/chromedp"
)

func ScrapeWishlist(wishlistID string) ([]WishlistItem, error) {
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
            return ScrollAndScrape(ctx, &items, uniqueItems)
        }),
    )

    if err != nil {
        return nil, fmt.Errorf("error in chromedp.Run: %v", err)
    }

    log.Printf("Scraping completed. Total unique items found: %d\n", len(items))
    return items, nil
}

func ScrollAndScrape(ctx context.Context, items *[]WishlistItem, uniqueItems map[string]bool) error {
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
