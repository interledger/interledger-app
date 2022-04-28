-- Create rafiki user and DB
CREATE DATABASE rafiki;
CREATE USER rafiki;
GRANT ALL ON DATABASE rafiki TO rafiki;

-- Create wallet user and DB. Represents an external wallet to which our rafiki can send.
CREATE DATABASE peer;
CREATE USER peer;
GRANT ALL ON DATABASE peer TO peer;
