-- Create an ADMIN user
CREATE USER IF NOT EXISTS roach WITH PASSWORD 'roach';
GRANT admin TO roach;

-- Create Kratos user and DB
CREATE DATABASE IF NOT EXISTS kratos;
CREATE USER IF NOT EXISTS kratos;
GRANT ALL ON DATABASE kratos TO kratos;