CREATE TABLE Profiles (
    Profiles_id_PK BIGINT PRIMARY KEY,
    Profiles_users_id_FK BIGINT,
    Profiles_correct_answers INT,
    Profiles_total_answers INT,
    Profiles_percentage INT
)