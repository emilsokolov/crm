package main

import (
	"database/sql"
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
	Product Product
	Sells   []Sell
}

type EditPageData struct {
	Product Product
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

	if r.Method == "POST" {
		quantityStr := r.FormValue("quantity")
		p, err := strconv.ParseInt(quantityStr, 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		quantity := int(p)

		err = product.Sell(quantity)
		if err == nil {
			saveProduct(product)
			saveSell(product.Id, quantity)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sells, err := getSells(product.Id)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	if r.Method == "GET" {
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
	} else if r.Method == "POST" {
		if r.FormValue("Save") == "Сохранить" {
			p, err := parseEditForm(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			p.Id = productID

			err = saveProduct(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/edit.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if r.FormValue("Cancel") == "Отменить" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		if r.FormValue("Save") == "Сохранить" {
			p, err := parseEditForm(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = addProduct(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}

	data := EditPageData{Product: Product{}}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseEditForm(r *http.Request) (Product, error) {
	name := r.FormValue("name")
	quantity := r.FormValue("quantity")
	sellprice := r.FormValue("sellprice")
	purchaseprice := r.FormValue("purchaseprice")

	quantityint, err := strconv.ParseInt(quantity, 10, 32)
	sellpriceint, err := strconv.ParseInt(sellprice, 10, 32)
	purchasepriceint, err := strconv.ParseInt(purchaseprice, 10, 32)

	if err != nil {
		return Product{}, err
	}

	return Product{
		Name:          name,
		Quantity:      int(quantityint),
		SellPrice:     int(sellpriceint),
		PurchasePrice: int(purchasepriceint),
	}, nil
}
