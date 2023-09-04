CREATE TABLE users(
    id serial PRIMARY KEY,
    login text,
    UNIQUE (login)
);
---- create above / drop below ----
DROP TABLE users;
