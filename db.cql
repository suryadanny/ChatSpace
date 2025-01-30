CREATE TABLE store.user (
    user_id text PRIMARY KEY,
    contact text,
    email text,
    last_active timestamp,
    name text,
    password text,
    user_name text
) ;

CREATE INDEX name_index ON store.user (user_name);


CREATE TABLE store.user_device (
    device_id text PRIMARY KEY,
    redis_id text,
    user_id text
);


CREATE TABLE store.chat_store (
    event_id uuid,
    sender_id text,
    received timestamp,
    delivered timestamp,
    is_delivered boolean,
    message text,
    receiver_id text,
    PRIMARY KEY ((event_id, sender_id), received)
) WITH CLUSTERING ORDER BY (received ASC);

