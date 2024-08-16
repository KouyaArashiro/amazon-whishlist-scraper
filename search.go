package main

import (
    "os"
    "io"
    "fmt"
    "log"
    "sync"
    "time"
    "strings"
    "net/url"
    "net/http"
    "io/ioutil"
    "database/sql"
    "encoding/json"

    _ "github.com/go-sql-driver/mysql"
)

//get isbn from wishlist_items
func SearchBooks(items []WishlistItem) (error) {
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


    if err != nil {
        log.Fatal(err)
    }

    results := processItems(items)
    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("Error (ISBN: %s): %v\n", result.ISBN, result.Error)
        } else {
            printBookInfo(result.BookInfo)
        }
        fmt.Println("---")
    }

    return nil
}

func processItems(items []WishlistItem) []Result {
    var wg sync.WaitGroup
    results := make([]Result, len(items))

    for i, item := range items {
        time.Sleep(5 * time.Second)
        wg.Add(1)
        go func(i int, item WishlistItem) {
            defer wg.Done()
            bookInfo, err := fetchBookInfo(item, 0)
            results[i] = Result{BookInfo: bookInfo, ISBN: item.ISBN, Error: err}
        }(i, item)
    }

    wg.Wait()
    return results
}

func fetchBookInfo(item WishlistItem, retryCount int) (*BookInfo, error) {
    baseURL := "https://api.calil.jp/check"
    appKey := os.Getenv("CALIL_APPKEY")
    systemID := "Univ_T_Kougei"
    format := "json"

    params := url.Values{}
    params.Add("appkey", appKey)
    params.Add("isbn", item.ISBN)
    params.Add("systemid", systemID)
    params.Add("format", format)

    fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

    resp, err := http.Get(fullURL)
    if err != nil {
        return nil, fmt.Errorf("Request Error: %v", err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Read response error: %v", err)
    }
    // レスポンスの内容をチェック
    if !strings.HasPrefix(string(body), "callback(") {
        return nil, fmt.Errorf("予期しないレスポンス形式: %s", truncateString(string(body), 100))
    }

    jsonStr := strings.TrimPrefix(string(body), "callback(")
    jsonStr = strings.TrimSuffix(jsonStr, ");")

    var response APIResponse
    err = json.Unmarshal([]byte(jsonStr), &response)
    if err != nil {
        return nil, fmt.Errorf("JSON perse error: %v", err)
    }

    if response.Continue == 1 {
        if retryCount < 3 {
            time.Sleep(5 * time.Second)
            return fetchBookInfo(item, retryCount+1)
        } else {
            return nil, fmt.Errorf("Exceed limit time")
        }
    }

    for _, libraryStatus := range response.Books[item.ISBN] {
        return &BookInfo{
            Title:      item.Title,
            ISBN:       item.ISBN,
            Status:     libraryStatus.Status,
            Libkey:     libraryStatus.Libkey,
            ReserveURL: libraryStatus.ReserveURL,
        }, nil
    }

    return nil, fmt.Errorf("Book info not found")
}

func printBookInfo(info *BookInfo) {
    logfile, err := os.OpenFile("./available.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        panic("cannnot open available.log:" + err.Error())
    }
    defer logfile.Close()

    log.SetOutput(io.MultiWriter(logfile, os.Stdout))
    log.SetFlags(log.Ldate | log.Ltime)

    if info.ReserveURL != "" {
        log.Printf("タイトル: %s\n", info.Title)
        log.Printf("ISBN: %s\n", info.ISBN)
        log.Printf("予約URL: %s\n", info.ReserveURL)
    }
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
