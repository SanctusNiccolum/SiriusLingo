CREATE TABLE Roles (
    Roles_id_PK BIGINT primary key,
    Roles_name TEXT,
    Roles_code INT,
    Roles_descr TEXT
);

CREATE TABLE Users (
	Users_id_PK BIGINT,
	Users_username VARCHAR(100),
	Users_password_hash TEXT,
	Users_email TEXT,
	Users_create TIMESTAMP DEFAULT now(),
	Users_auth TIMESTAMP DEFAULT now(),
	Users_roles_id_FK BIGINT,
	FOREIGN KEY(Users_roles_id_FK) references Roles(Roles_id_PK)
)