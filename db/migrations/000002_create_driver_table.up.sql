CREATE TABLE IF NOT EXISTS drivers(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    loc integer DEFAULT -100,
    available BOOLEAN NOT NULL DEFAULT TRUE,
    rating real DEFAULT 5.0,
    with_passenger DEFAULT NULL,
    last_trip_end_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
