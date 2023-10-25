create table products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    purchase_price INTEGER NOT NULL,
    sell_price INTEGER NOT NULL,
    create_date TEXT NOT NULL,
    update_date TEXT NOT NULL
);

create table sells (
    product_id INTEGER NOT NULL,
    sell_date TEXT NOT NULL,
    quantity INTEGER NOT NULL
);

insert into products (name, quantity, purchase_price, sell_price, create_date, update_date)
values
    ('Банка 0.5', 50, 15, 30, '2023-10-13 15:49:00', '2023-10-13 15:49:00')
    ,('Банка 1', 17, 20, 40, '2023-10-13 15:50:00', '2023-10-13 15:50:00')
    ,('Банка 3', 2, 25, 50, '2023-10-13 15:51:20', '2023-10-13 15:51:20')
    ,('Крышка красная', 100, 3, 5, '2023-10-13 15:52:30', '2023-10-13 15:52:30')
    ,('Перчатки', 25, 13, 25, '2023-10-13 15:53:24', '2023-10-13 15:53:24');

insert into sells(product_id, sell_date, quantity)
values
(1, '2023-10-13 16:00', 5),
(2, '2023-10-13 16:10', 2),
(3, '2023-10-13 16:08', 1),
(4, '2023-10-13 16:24', 15),
(5, '2023-10-13 16:13', 3);
