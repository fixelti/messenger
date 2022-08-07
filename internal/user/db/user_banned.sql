CREATE TABLE users
(
    id serial primary key,
    created_at timestamp DEFAULT current_timestamp,
    deleted_at timestamp,
    user_id serial,
    banned_user_id serial
);