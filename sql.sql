CREATE TABLE Customers (
                           ID SERIAL PRIMARY KEY,
                           first_name           VARCHAR(100) NOT NULL,
                           last_name            VARCHAR(100) NOT NULL,
                           birth_date         DATE DEFAULT '1900-01-01' NOT NULL,
                           gender             VARCHAR(6) NOT NULL  CHECK ( gender = 'Male' OR gender = 'Female'),
                           email        VARCHAR(100) NOT NULL,
                           address           VARCHAR(100)
);

/*just execute this many times to test search*/
INSERT INTO Customers(first_name,  last_name, birth_date, gender, email, address)
VALUES ('Aleksei', 'Sinitsyn', '1992-06-28', 'Male', 'mail@gmail.com', 'tetxetetxetetxet'),
       ('Irina', 'Ivanova', '1994-07-03', 'Female', 'gmail@mail.com', 'testtextetxetxetx');
