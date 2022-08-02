CREATE TABLE IF NOT EXISTS passengers(
    name VARCHAR(100) NOT NULL PRIMARY KEY,
    password VARCHAR(100) NOT NULL,
    role VARCHAR(100),
    workflow_id VARCHAR(100),
    rating real DEFAULT 5.0
);
