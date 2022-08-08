CREATE TABLE IF NOT EXISTS passengers(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    pick_up_loc integer DEFAULT -100,
    drop_loc integer DEFAULT -100,
    rating real DEFAULT 5.0,
    workflow_id VARCHAR(100),
    in_ride BOOLEAN NOT NULL DEFAULT FALSE,
    with_driver integer DEFAULT -1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
