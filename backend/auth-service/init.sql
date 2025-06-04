DROP TABLE users;
DROP TABLE roles;

CREATE TABLE roles (
   roles_id_pk BIGSERIAL PRIMARY KEY,
   roles_name TEXT UNIQUE,
   roles_code INT UNIQUE,
   roles_descr TEXT
);

CREATE TABLE users (
   users_id_pk BIGSERIAL PRIMARY KEY,
   users_username VARCHAR(100) UNIQUE,
   users_password_hash TEXT,
   users_email TEXT UNIQUE,
   users_auth_time TIMESTAMP DEFAULT now(),
   users_roles_id_fk BIGINT,
   users_access_token_secret TEXT,
   users_refresh_token_secret TEXT,
   users_access_token_jti UUID,
   users_refresh_token_jti UUID,
   users_created_at TIMESTAMP DEFAULT now(),
   users_updated_at TIMESTAMP DEFAULT now(),
   FOREIGN KEY(users_roles_id_fk) REFERENCES roles(roles_id_pk)
);

INSERT INTO roles (roles_name, roles_code, roles_descr)
VALUES ('user', 1, 'default user of app');