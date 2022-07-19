CREATE TABLE IF NOT EXISTS drivers(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    loc integer,
    available BOOLEAN NOT NULL DEFAULT TRUE,
    rating real,
    with_passenger DEFAULT NULL,
    last_trip_end_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
