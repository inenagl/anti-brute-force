CREATE DATABASE abf;
CREATE USER abfuser WITH ENCRYPTED PASSWORD 'abfpassword';
GRANT ALL ON DATABASE abf TO abfuser;
ALTER DATABASE abf OWNER TO abfuser;
GRANT USAGE, CREATE ON SCHEMA PUBLIC TO abfuser;
