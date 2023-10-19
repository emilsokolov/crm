package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type ProductPageData struct {
	Title    string
	Products []Product
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "crm.db")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/products/", productHandler)
	http.HandleFunc("/product/", editHandler)
	log.Println("Crm started...")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts()
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.ParseFiles("products.html")
	if err != nil {
		log.Fatal(err)
	}

	data := ProductPageData{
		Title:    "Все товары",
		Products: products,
	}

	t.Execute(w, data)
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.URL.Path[len("/products/"):]
	i, err := strconv.ParseInt(idString, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	id := int(i)
	product, err := getProduct(id)
	if err != nil {
		if err.Error() == "Product Not Found!" {
			fmt.Fprintf(w, "Product Not Found")
		} else {
			log.Fatal(err)
		}
	}
	fmt.Fprintf(w, "Наименование = %s, Остаток = %d, Закупка: %d, Продажа: %d",
		product.Name, product.Quantity, product.PurchasePrice, product.SellPrice)
}

func editHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Edit")
}
