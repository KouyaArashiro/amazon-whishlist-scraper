package main

import (
    "database/sql"
    "os"
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql"
)

func SaveToDatabase(items []WishlistItem) error {
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
    existingTitles, err := GetExistingTitles(tx)
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

func GetExistingTitles(tx *sql.Tx) (map[string]bool, error) {
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
