package main

import (
	"database/sql"
	"fmt"
	"log"

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

	fmt.Println(products)

	fmt.Printf("%2s\t%20s\t%s\n", "Id", "Name", "Quantity")
	for _, p := range products {
		fmt.Printf("%2d\t%20s\t%3d\n", p.Id, p.Name, p.Quantity)
	}
}
