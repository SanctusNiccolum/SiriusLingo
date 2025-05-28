CREATE TABLE Languages(
    Languages_id_PK BIGINT PRIMARY KEY,
    Languages_name VARCHAR(100),
    Languages_abr VARCHAR(5)
);

CREATE TABLE Words(
    Words_id_PK BIGINT,
    Words_text VARCHAR(100),
    Words_languages_id_FK BIGINT,
    Words_groups_id BIGINT,
    Words_etimology TEXT,
    FOREIGN KEY(Words_languages_id_FK) references Languages(Languages_id_PK)
)