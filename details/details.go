package details

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ServiceWeaver/weaver"
)

type BookDetails struct {
	weaver.AutoMarshal
	ID        int    `json:"id"`
	Author    string `json:"author"`
	Year      int    `json:"year"`
	Type      string `json:"type"`
	Pages     int    `json:"pages"`
	Publisher string `json:"publisher"`
	Language  string `json:"language"`
	ISBN10    string `json:"ISBN-10"`
	ISBN13    string `json:"ISBN-13"`
}

type Details interface {
	GetBookDetails(context.Context, int, map[string]string) (BookDetails, error)
}

type details struct {
	weaver.Implements[Details]
}

func (d *details) GetBookDetails(_ context.Context, id int, headers map[string]string) (BookDetails, error) {
	if os.Getenv("ENABLE_EXTERNAL_BOOK_SERVICE") == "true" {
		isbn := "0486424618"
		return fetchDetailsFromExternalService(isbn, id, headers)
	}

	return BookDetails{
		ID:        id,
		Author:    "William Shakespeare",
		Year:      1595,
		Type:      "paperback",
		Pages:     200,
		Publisher: "PublisherA",
		Language:  "English",
		ISBN10:    "1234567890",
		ISBN13:    "123-1234567890",
	}, nil
}

func fetchDetailsFromExternalService(isbn string, id int, headers map[string]string) (BookDetails, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := "https://www.googleapis.com/books/v1/volumes?q=isbn:" + isbn

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return BookDetails{}, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return BookDetails{}, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return BookDetails{}, err
	}

	book := result["items"].([]interface{})[0].(map[string]interface{})["volumeInfo"].(map[string]interface{})

	language := book["language"].(string)
	if language == "en" {
		language = "English"
	} else {
		language = "unknown"
	}

	bookType := book["printType"].(string)
	if bookType == "BOOK" {
		bookType = "paperback"
	} else {
		bookType = "unknown"
	}

	isbn10 := getISBN(book, "ISBN_10")
	isbn13 := getISBN(book, "ISBN_13")

	yearStr := book["publishedDate"].(string)
	year, err := strconv.Atoi(yearStr[:4])
	if err != nil {
		log.Printf("Failed to extract year: %v", err)
		year = 0
	}

	return BookDetails{
		ID:        id,
		Author:    book["authors"].([]interface{})[0].(string),
		Year:      year,
		Type:      bookType,
		Pages:     int(book["pageCount"].(float64)),
		Publisher: book["publisher"].(string),
		Language:  language,
		ISBN10:    isbn10,
		ISBN13:    isbn13,
	}, nil
}

func getISBN(book map[string]interface{}, isbnType string) string {
	for _, identifier := range book["industryIdentifiers"].([]interface{}) {
		id := identifier.(map[string]interface{})
		if id["type"] == isbnType {
			return id["identifier"].(string)
		}
	}
	return ""
}
