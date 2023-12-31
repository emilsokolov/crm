package main

import (
	"errors"
	"time"
)

type Product struct {
	Id            int
	Name          string
	Quantity      int
	PurchasePrice int
	SellPrice     int
}

func (p *Product) Sell(quantity int) error {
	if quantity < 0 {
		return errors.New("quantity must be positive")
	}
	if quantity > p.Quantity {
		return errors.New("not enough quantity")
	}
	p.Quantity -= quantity
	return nil
}

func saveProduct(p Product) error {
	updateDate := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(`
update products
set
	name = ?,
	sell_price = ?,
	purchase_price = ?,
	quantity = ?,
	update_date = ?
where id = ?;`, p.Name, p.SellPrice, p.PurchasePrice, p.Quantity, updateDate, p.Id)
	return err
}

func deleteProduct(productID int) error {
	_, err := db.Exec(`
delete from products
where id = ?;`, productID)
	return err
}

func saveSell(productId, quantity int) error {
	sellDate := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec("insert into sells(product_id, sell_date, quantity) values (?,?,?);", productId, sellDate, quantity)
	return err
}

func getProducts() ([]Product, error) {
	rows, err := db.Query("select id, name, quantity, purchase_price, sell_price from products;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		p := Product{}
		err = rows.Scan(&p.Id, &p.Name, &p.Quantity, &p.PurchasePrice, &p.SellPrice)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func getProduct(id int) (Product, error) {
	p := Product{}
	rows, err := db.Query("select id, name, quantity, purchase_price, sell_price from products where id = ?;", id)
	if err != nil {
		return p, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&p.Id, &p.Name, &p.Quantity, &p.PurchasePrice, &p.SellPrice)
	} else {
		err = errors.New("product not found")
	}
	return p, err
}

func getSells(productId int) ([]Sell, error) {
	rows, err := db.Query("select product_id, sell_date, quantity from sells where product_id = ? order by sell_date desc;", productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sells := []Sell{}
	for rows.Next() {
		var s Sell
		err = rows.Scan(&s.ProductId, &s.Date, &s.Quantity)
		if err != nil {
			return nil, err
		}
		sells = append(sells, s)
	}
	return sells, nil
}

func addProduct(product Product) error {
	addDate := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec("insert into products(name,purchase_price, sell_price, quantity, create_date, update_date) values (?,?,?,?,?,?);",
		product.Name, product.PurchasePrice, product.SellPrice, product.Quantity, addDate, addDate)
	return err
}
