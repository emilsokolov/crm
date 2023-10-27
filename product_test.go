package main

// func TestProductSell(t *testing.T) {
// 	testData := []struct {
// 		Quantity int
// 		Error    string
// 	}{
// 		{1, ""},
// 		{0, "quantity must be positive"},
// 	}

// 	for _, d := range testData {
// 		p := Product{Quantity: 10}
// 		err := p.Sell(d.Quantity)

// 		if d.Error == "" && err != nil || d.Error != "" && err.Error() {
// 			t.Errorf("want error = %s, but was = %v", d.Error, err)
// 		}
// 	}
// }
