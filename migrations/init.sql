CREATE TABLE users
(
    username      VARCHAR PRIMARY KEY,
    password_hash VARCHAR NOT NULL,
    balance       INTEGER
);

CREATE TABLE items
(
    type  VARCHAR PRIMARY KEY,
    price INTEGER
);

INSERT INTO items
VALUES ('t-shirt', 80),
       ('cup', 20),
       ('book', 50),
       ('pen', 10),
       ('powerbank', 200),
       ('hoody', 300),
       ('umbrella', 200),
       ('socks', 10),
       ('wallet', 50),
       ('pink-hoody', 500);

CREATE TABLE transactions
(
    sender     VARCHAR REFERENCES users (username),
    receiver   VARCHAR REFERENCES users (username),
    amount     INTEGER,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE purchases
(
    username VARCHAR REFERENCES users (username),
    item     VARCHAR REFERENCES items (type),
    quantity INTEGER,
    CONSTRAINT username_item UNIQUE (username, item)
)