package main

import (
	"errors"
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

func saveProduct(p *Product) error {
	_, err := db.Exec("update products set quantity = ? where id = ?;", p.Quantity, p.Id)
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
	rows, err := db.Query("select product_id, sell_date, quantity from sells where product_id = ?;", productId)
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
