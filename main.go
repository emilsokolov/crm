package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type ProductsPageData struct {
	Title    string
	Products []Product
}

type ProductPageData struct {
	Product       Product
	Sells         []Sell
	Quantity      string
	QuantityError string
}

type EditPageData struct {
	Product            Product
	NameError          string
	QuantityError      string
	PurchasePriceError string
	SellPriceError     string
	ParseError         error
}

type Sell struct {
	Date      string
	Quantity  int
	ProductId int
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "db/crm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/{$}", rootHandler)
	http.HandleFunc("/styles.css", cssHandler)
	http.HandleFunc("/favicon.ico", icoHandler)
	http.HandleFunc("/products/{id}", productHandler)
	http.HandleFunc("/products/new", newHandler)
	http.HandleFunc("/products/{id}/edit", editHandler)
	http.HandleFunc("POST /products/{id}/delete", deleteHandler)
	log.Println("Crm started at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
	w.Header().Add("Content-Type", "text/css")
	w.Write(f)
}

func icoHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.ReadFile("templates/favicon.ico")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Type", "image/x-icon")
	w.Write(f)
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	productID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Print("productHander: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	product, err := getProduct(int(productID))
	if err != nil {
		log.Print("productHander: ", err)

		if err.Error() == "product not found" {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	var data ProductPageData

	if r.Method == "POST" {
		quantity, err := parseQuantity(r.FormValue("quantity"))
		if err == nil {
			err = product.Sell(quantity)
			if err == nil {
				saveProduct(product)
				saveSell(product.Id, quantity)
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		data.Quantity = r.FormValue("quantity")
		data.QuantityError = err.Error()
	}

	sells, err := getSells(product.Id)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data.Product = product
	data.Sells = sells

	t, err := template.ParseFiles("templates/product.html")
	if err != nil {
		log.Fatal(err)
	}

	t.Execute(w, data)
}

func parseQuantity(quantityStr string) (int, error) {
	if quantityStr == "" {
		return 0, fmt.Errorf("Поле не должно быть пустым")
	}

	p, err := strconv.ParseInt(quantityStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}

	quantity := int(p)
	if quantity <= 0 {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}

	return quantity, nil
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	productID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Print("editHandler: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data EditPageData

	if r.Method == "GET" {
		product, err := getProduct(int(productID))
		if err != nil {
			log.Print("editHander: getProduct: ", err)

			if err.Error() == "product not found" {
				http.NotFound(w, r)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}
		data.Product = product
	} else if r.Method == "POST" {
		data, err = parseEditForm(r)
		if err == nil {
			data.Product.Id = int(productID)

			err = saveProduct(data.Product)
			if err != nil {
				log.Print("editHander: saveProduct: ", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	t, err := template.ParseFiles("templates/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var b strings.Builder
	err = t.Execute(&b, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(b.String()))
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	productID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	product, err := getProduct(int(productID))
	if err != nil {
		log.Print("deleteHandler: ", err)

		if err.Error() == "product not found" {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	err = deleteProduct(product.Id)
	if err != nil {
		log.Print("deleteHandler: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	var data EditPageData

	if r.Method == "POST" {
		var err error
		data, err = parseEditForm(r)
		if err == nil {
			err = addProduct(data.Product)
			if err != nil {
				log.Print("newHander: addProduct: ", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	t, err := template.ParseFiles("templates/edit.html")
	if err != nil {
		log.Print("newHander: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Print("newHander: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func parseEditForm(r *http.Request) (EditPageData, error) {
	var data EditPageData

	name := r.FormValue("name")
	if name == "" {
		data.NameError = "Имя не заполнено"
		data.ParseError = fmt.Errorf("name is empty")
	}

	quantity := r.FormValue("quantity")
	if quantity == " " {
		data.QuantityError = "Поле не должно быть пустым"
	}
	sellprice := r.FormValue("sellprice")
	if sellprice == " " {
		data.SellPriceError = "Поле не должно быть пустым"
	}
	purchaseprice := r.FormValue("purchaseprice")
	if purchaseprice == " " {
		data.PurchasePriceError = "Поле не должно быть пустым"
	}

	quantityint, err := parseQuantityInt(quantity)
	if err != nil {
		data.ParseError = err
	}
	sellpriceint, err := parseSellPriceInt(sellprice)
	if err != nil {
		data.ParseError = err
	}
	purchasepriceint, err := parsePurchasePriceInt(purchaseprice)
	if err != nil {
		data.ParseError = err
	}
	if err == nil {
		data.Product = Product{
			Name:          name,
			Quantity:      int(quantityint),
			SellPrice:     int(sellpriceint),
			PurchasePrice: int(purchasepriceint),
		}
	}

	return data, data.ParseError
}
func parseQuantityInt(quantityint string) (int, error) {
	p, err := strconv.ParseInt(quantityint, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}
	quantity := int(p)
	if quantity <= 0 {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}

	return quantity, nil
}
func parseSellPriceInt(sellpriceint string) (int, error) {

	p, err := strconv.ParseInt(sellpriceint, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}

	sellprice := int(p)
	if sellprice <= 0 {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}

	return sellprice, nil
}
func parsePurchasePriceInt(purchasepriceint string) (int, error) {
	p, err := strconv.ParseInt(purchasepriceint, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}

	purchaseprice := int(p)
	if purchaseprice <= 0 {
		return 0, fmt.Errorf("Ожидается целое положительное число")
	}

	return purchaseprice, nil
}
