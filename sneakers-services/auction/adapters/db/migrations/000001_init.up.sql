CREATE TABLE auctions (
    id SERIAL PRIMARY KEY,
    sneaker_id INT NOT NULL,
    current_price BIGINT NOT NULL,
    end_at TIMESTAMP NOT NULL
);

CREATE TABLE bids (
    id SERIAL PRIMARY KEY,
    auction_id INT REFERENCES auctions(id) ON DELETE CASCADE,
    user_id INT NOT NULL,
    amount BIGINT NOT NULL
);