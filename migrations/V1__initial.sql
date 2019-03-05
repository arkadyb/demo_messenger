CREATE SEQUENCE message_id_seq;
CREATE TABLE messages (
    message_id bigint NOT NULL DEFAULT nextval('message_id_seq') PRIMARY KEY,
    originator text NOT NULL,
    "text" text NOT NULL,
    processed boolean NOT NULL DEFAULT FALSE
);

CREATE TABLE recipients (
    message_id bigint NOT NULL,
    phone_number text NOT NULL
);
CREATE UNIQUE INDEX msgid_phonenumber_idx ON recipients(message_id ,phone_number);
