package main

type WishlistItem struct {
    Title string
    Price string
    URL   string
    ISBN  string
}

type LibraryStatus struct {
    Status     string            `json:"status"`
    Libkey     map[string]string `json:"libkey"`
    ReserveURL string            `json:"reserveurl"`
}

type APIResponse struct {
    Session  string                           `json:"session"`
    Continue int                              `json:"continue"`
    Books    map[string]map[string]LibraryStatus `json:"books"`
}

type Result struct {
    BookInfo *BookInfo
    ISBN     string
    Error    error
}

type BookInfo struct {
    Title      string
    ISBN       string
    Status     string
    Libkey     map[string]string
    ReserveURL string
}
