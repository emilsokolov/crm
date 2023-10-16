package main

import "fmt"

type Product struct {
	Id            int
	Name          string
	Quantity      int
	PurchasePrice int
	SellPrice     int
}

func getProducts() ([]Product, error) {
	rows, err := db.Query("select id, name, quantity, purchase_price, sell_price from products;")
	if err != nil {
		return nil, err
	}
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
	if rows.Next() {
		err = rows.Scan(&p.Id, &p.Name, &p.Quantity, &p.PurchasePrice, &p.SellPrice)
	} else {
		err = fmt.Errorf("product with id = %d not found", id)
	}
	return p, err
}
