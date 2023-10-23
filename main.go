package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type ProductsPageData struct {
	Title    string
	Products []Product
}

type ProductPageData struct {
	Product Product
	Sells   []Sell
}

type Sell struct {
	Date      string
	Quantity  int
	ProductId int
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "crm.db")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/styles.css", cssHandler)
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

	t, err := template.ParseFiles("templates/products.html")
	if err != nil {
		log.Fatal(err)
	}

	data := ProductsPageData{
		Title:    "Все товары",
		Products: products,
	}

	t.Execute(w, data)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.ReadFile("templates/styles.css")
	if err != nil {
		log.Fatal(err)
	}
	w.Write(f)
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
	sells, err := getSells(product.Id)
	if err != nil {
		log.Fatal(err)
	}

	data := ProductPageData{
		Product: product,
		Sells:   sells,
	}

	t, err := template.ParseFiles("templates/product.html")
	if err != nil {
		log.Fatal(err)
	}

	t.Execute(w, data)
}

func editHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Edit")
}
