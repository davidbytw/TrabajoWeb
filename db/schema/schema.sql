CREATE TABLE books (
    id        SERIAL PRIMARY KEY,
    title     VARCHAR(255) NOT NULL,
    author    VARCHAR(255) NOT NULL,
    status    VARCHAR(20)  NOT NULL
              CHECK (status IN ('no_leido', 'leyendo', 'leido')),
    read_date DATE,
    rating    INTEGER CHECK (rating BETWEEN 1 AND 5),
    notes     TEXT
);