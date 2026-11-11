-- las reglas tambien se validan en logic: estos CHECK son la segunda barrera
CREATE TABLE book (
    id        SERIAL PRIMARY KEY,
    title     VARCHAR(255) NOT NULL,
    author    VARCHAR(255) NOT NULL,
    status    VARCHAR(20)  NOT NULL DEFAULT 'no_leido',
    read_date DATE,
    rating    INTEGER,
    notes     TEXT,

    CONSTRAINT book_title_not_blank  CHECK (btrim(title) <> ''),
    CONSTRAINT book_author_not_blank CHECK (btrim(author) <> ''),
    CONSTRAINT book_status_valid     CHECK (status IN ('no_leido', 'leyendo', 'leido')),
    CONSTRAINT book_rating_range     CHECK (rating BETWEEN 1 AND 5),
    -- leido exige fecha y puntaje; los demas estados no llevan fecha, puntaje ni notas
    CONSTRAINT book_finished_fields CHECK (
        (status = 'leido' AND read_date IS NOT NULL AND rating IS NOT NULL)
        OR
        (status <> 'leido' AND read_date IS NULL AND rating IS NULL AND notes IS NULL)
    )
);
