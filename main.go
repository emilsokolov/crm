package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type product struct {
	Id       int
	Name     string
	Quantity int
}

func main() {
	db, err := sql.Open("sqlite3", "crm.db")

	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query("select id, name, quantity from products;")
	if err != nil {
		log.Fatal(err)
	}
	var products []product
	for rows.Next() {
		p := product{}
		err = rows.Scan(&p.Id, &p.Name, &p.Quantity)
		if err != nil {
			log.Fatal(err)
		}
		products = append(products, p)
	}

	http.HandleFunc("/", productsHandler)
	log.Println("Crm started...")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}
