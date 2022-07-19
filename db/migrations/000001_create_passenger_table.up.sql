CREATE TABLE IF NOT EXISTS passengers(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    pick_up_loc integer,
    drop_loc integer,
    rating real,
    in_ride BOOLEAN NOT NULL DEFAULT FALSE,
    with_driver DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
