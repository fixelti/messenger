CREATE TABLE IF NOT EXISTS users
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
    friends integer[],
    user_role integer
);

CREATE TABLE IF NOT EXISTS users_banned
(
    id serial primary key,
    created_at timestamp DEFAULT current_timestamp,
    deleted_at timestamp,
    user_id serial,
    banned_user_id serial
);

INSERT INTO users(id, email, login, password, secret_word, find_vision, add_friend, friends, user_role)
VALUES(1,'root', 'rootOleg', '$2a$04$d/HBp21BclQ/FXxCcKILMOg5CEGrHP0.emPTNQIgGurBHtLJcaFnu', 'root', false, false, '{0}', 1);