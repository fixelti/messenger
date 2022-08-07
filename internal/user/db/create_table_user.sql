CREATE TABLE users
(
    id serial primary key,
    created_at timestamp DEFAULT current_timestamp,
    deleted_at timestamp,
    email varchar(100),
    login varchar(100),
    password varchar(100),
    secret_word varchar(100),
    find_vision boolean DEFAULT false,
    add_friend boolean DEFAULT false,
    friends []integer,
);