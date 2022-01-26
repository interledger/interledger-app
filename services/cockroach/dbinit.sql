-- Create an ADMIN user
CREATE USER IF NOT EXISTS roach WITH PASSWORD 'roach';
GRANT admin TO roach;

-- Create Kratos user and DB
CREATE DATABASE IF NOT EXISTS kratos;
CREATE USER IF NOT EXISTS kratos;
GRANT ALL ON DATABASE kratos TO kratos;

-- Create backend user and DB
CREATE DATABASE IF NOT EXISTS backend;
CREATE USER IF NOT EXISTS backend;
GRANT ALL ON DATABASE backend TO backend;

-- Create pacioli user and DB
CREATE DATABASE IF NOT EXISTS pacioli;
CREATE USER IF NOT EXISTS pacioli;
GRANT ALL ON DATABASE pacioli TO pacioli;