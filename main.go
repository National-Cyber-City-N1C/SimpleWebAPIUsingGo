//Sharing by N1C

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

//Before you test, you should already have MySQL or Mariadb.
//If you already have these database, create database called go_crud.
//Use the database called go_crud and then create table called book_lists
//Query to create table look like below
//CREATE TABLE book_lists (
//	id INT AUTO_INCREMENT PRIMARY KEY,
//  name VARCHAR(50) NOT NULL,
//	author VARCHAR(50) NOT NULL,
//	price DOUBLE NOT NULL
//);

func initDb() {
	var err error
	//replace with your (MySQL or Mariadb) database
	//                       username:password	  host	  port   dbbane
	db, err = sql.Open("mysql", "root:password@(127.0.0.1:3306)/go_crud?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success database connection")
}

func main() {
	initDb()

	r := mux.NewRouter()

	r.HandleFunc("/api/books", getAllBooks).Methods("GET")
	r.HandleFunc("/api/books/{book_id}", getOneBook).Methods("GET")
	r.HandleFunc("/api/books", createBook).Methods("POST")
	r.HandleFunc("/api/books/{book_id}", updateBook).Methods("PUT")
	r.HandleFunc("/api/books/{book_id}", deleteBook).Methods("DELETE")

	fmt.Println("Server run on http://localhost:8080")

	http.ListenAndServe(":8080", r)

}

func getAllBooks(w http.ResponseWriter, r *http.Request) {
	type Book struct {
		Id     int
		Name   string
		Author string
		Price  float64
	}
	rows, err := db.Query(`SELECT id, name, author, price FROM book_lists`)
	if err != nil {
		fmt.Println("Error: select all books")
		return
	}
	defer rows.Close()
	var books []Book
	for rows.Next() {
		var b Book
		err := rows.Scan(&b.Id, &b.Name, &b.Author, &b.Price)
		if err != nil {
			fmt.Println("Error: scanning all books")
			return
		}
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Error: query all rows")
		return
	}
	json.NewEncoder(w).Encode(books)
}

func getOneBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	book_id := vars["book_id"]

	var (
		id     int
		name   string
		author string
		price  float64
	)

	query := `SELECT id, name, author, price FROM book_lists WHERE id = ?`
	db.QueryRow(query, book_id).Scan(&id, &name, &author, &price)

	bookData := map[string]interface{}{
		"id":     id,
		"name":   name,
		"author": author,
		"price":  price,
	}

	json.NewEncoder(w).Encode(bookData)

}

func createBook(w http.ResponseWriter, r *http.Request) {
	var newBook struct {
		Name   string  `json:"name"`
		Author string  `json:"author"`
		Price  float64 `json:"price"`
	}

	err := json.NewDecoder(r.Body).Decode(&newBook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO book_lists (name, author, price) VALUES (?, ?, ?)", newBook.Name, newBook.Author, newBook.Price)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error creating the book", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error getting the book ID", http.StatusInternalServerError)
		return
	}

	createdBook := map[string]interface{}{
		"id":     id,
		"name":   newBook.Name,
		"author": newBook.Author,
		"price":  newBook.Price,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdBook)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["book_id"]

	var updatedBook struct {
		Name   string  `json:"name"`
		Author string  `json:"author"`
		Price  float64 `json:"price"`
	}

	err := json.NewDecoder(r.Body).Decode(&updatedBook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE book_lists SET name = ?, author = ?, price = ? WHERE id = ?", updatedBook.Name, updatedBook.Author, updatedBook.Price, bookID)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error updating the book", http.StatusInternalServerError)
		return
	}

	successMessage := map[string]string{"message": "Book updated successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(successMessage)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["book_id"]

	_, err := db.Exec("DELETE FROM book_lists WHERE id = ?", bookID)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error deleting the book", http.StatusInternalServerError)
		return
	}

	successMessage := map[string]string{"message": "Book deleted successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(successMessage)
}
