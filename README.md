# Amazon Wishlist Scraper
Amazon Wishlist Scraper is a tool that scrapes book information from an Amazon wishlist, then uses the retrieved ISBNs to check the availability of these books in specified libraries. The scraped data is saved to a database, and any available books are logged into a file.

# Features
Wishlist Scraping: Extracts book titles, prices, URLs, and ISBNs from an Amazon wishlist.
Database Storage: Saves the scraped data into a MySQL database while avoiding duplicates.
Library Availability Check: Uses the Calil Library API to check the availability of books in specified libraries.
Logging: Records the reservation URLs of available books into a log file.
# Prerequisites
Go programming language (version 1.16 or higher)
MySQL database
Calil Library API key
Google Chrome and ChromeDriver
Environment variables setup (see Configuration for details)

# Installation
Clone the repository:

```
git clone https://github.com/KouyaArashiro/amazon-whishlist-scraper.git
cd amazon-whishlist-scraper
go get
```

# Configuration
Create a .env file in the project root directory and set the following environment variables:

```
AMAZON_WISHLIST_ID=Your Amazon Wishlist ID
DB_USER=Your database username
DB_PASS=Your database password
DB_NAME=Your database name
DB_HOST=Your database host (e.g., localhost)
CALIL_APPKEY=Your Calil API key
Database Setup
Use the following SQL query to create the wishlist_items table:
```
sql
```
CREATE TABLE wishlist_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255),
    price VARCHAR(50),
    url TEXT,
    isbn VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Usage
Set up the .env file with the necessary environment variables.

Run the program:

bash
```
go run main.go
```
The program will execute the following steps:

Scrape book information from your Amazon wishlist.
Save the unique book data to the database.
Check the availability of these books in the specified library using the Calil Library API.
Log any available books' reservation URLs into the available.log file.
Log File
When available books are found, they are recorded in the available.log file in the following format:
```
YYYY/MM/DD HH:MM:SS Title: Book Title
YYYY/MM/DD HH:MM:SS ISBN: Book ISBN
YYYY/MM/DD HH:MM:SS Reservation URL: Library Reservation URL
```
# Notes
Compliance with Terms of Service: Please ensure you use this tool in accordance with Amazon's and Calil Library API's terms of service.
Scraping Etiquette: The tool includes intervals to prevent excessive load on the scraped websites. Adjust these intervals responsibly if necessary.
Security: Be cautious with sensitive information like environment variables and API keys to prevent unauthorized access.
License
This project is licensed under the MIT License.

# Contributing
Contributions such as bug reports, feature suggestions, and pull requests are welcome.

# Author
Kouya Arashiro  
Contact  
GitHub: KouyaArashiro  
# Details
Below is an explanation of the main functionalities and code structure of this tool.

## main.go
Loads environment variables to obtain the wishlist ID, database connection info, and API keys.
Calls the ScrapeWishlist function to get book information from the wishlist.
Saves the retrieved data to the database using the SaveToDatabase function.
Checks the library availability by calling the SearchBooks function.
## scraper.go
The ScrapeWishlist function uses Chromedp to scrape the Amazon wishlist page.
It scrolls through the page to load all items and extracts book information, including the ISBN.
Filters out duplicates and items without an ISBN.
## db.go
Handles database connections and operations.
The SaveToDatabase function inserts new book data into the wishlist_items table.
Checks for existing titles to prevent duplicate entries.
## search.go
The SearchBooks function checks the availability of books using the Calil Library API.
Fetches book information concurrently for efficiency.
Outputs the results to a log file.
## models.go
Defines the data structures used in the program.
Includes structs like WishlistItem, LibraryStatus, APIResponse, Result, and BookInfo.
## Future Improvements
Enhanced Error Handling: Improve handling of network and API errors.
Unit Testing: Implement tests for better reliability.
Configuration Flexibility: Allow more flexible settings via command-line arguments or configuration files.
If you have any questions or issues, please feel free to contact me.
