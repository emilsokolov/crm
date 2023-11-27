package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

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

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/styles.css", cssHandler)
	http.HandleFunc("/products/", productsHandler)
	http.HandleFunc("/products/new", newHandler)
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

var productValidPath = regexp.MustCompile("^/products/([0-9]+)(/edit)?$")

func productsHandler(w http.ResponseWriter, r *http.Request) {
	m := productValidPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}

	productID, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, "Oops...", http.StatusInternalServerError)
		return
	}

	if m[2] == "/edit" {
		editHandler(w, r, int(productID))
		return
	}

	productHandler(w, r, int(productID))
}

func productHandler(w http.ResponseWriter, r *http.Request, productID int) {
	product, err := getProduct(productID)
	if err != nil {
		log.Print(err)

		if err.Error() == "product not found" {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	data := ProductPageData{}

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

func editHandler(w http.ResponseWriter, r *http.Request, productID int) {
	product, err := getProduct(productID)
	if err != nil {
		log.Print(err)

		if err.Error() == "product not found" {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	if r.Method == "POST" && r.FormValue("Save") == "Сохранить" {
		data, err := parseEditForm(r)
		if err == nil {
			data.Product.Id = productID

			err = saveProduct(data.Product)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}

	t, err := template.ParseFiles("templates/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := EditPageData{Product: product}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	data := EditPageData{}

	if r.Method == "POST" {
		data, err = parseEditForm(r)
		if err == nil {
			err = addProduct(data.Product)
			if err != nil {
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

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseEditForm(r *http.Request) (EditPageData, error) {
	data := EditPageData{}

	name := r.FormValue("name")
	if name == "" {
		data.NameError = "Имя не заполнено"
		data.ParseError = fmt.Errorf("name is empty")
	}

	quantity := r.FormValue("quantity")
	sellprice := r.FormValue("sellprice")
	purchaseprice := r.FormValue("purchaseprice")

	quantityint, err := strconv.ParseInt(quantity, 10, 32)
	if err != nil {
		data.ParseError = err
	}

	sellpriceint, err := strconv.ParseInt(sellprice, 10, 32)
	if err != nil {
		data.ParseError = err
	}

	purchasepriceint, err := strconv.ParseInt(purchaseprice, 10, 32)
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
