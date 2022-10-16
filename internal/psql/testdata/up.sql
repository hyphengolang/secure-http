CREATE TABLE testcase (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    age INT NOT NULL CHECK (age > 0)
);