CREATE TABLE IF NOT EXISTS Users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR (100) UNIQUE NOT NULL,
    hashed_password VARCHAR (100) NOT NULL,
    first_name VARCHAR(200),
    phone_number VARCHAR(15),
    user_address VARCHAR(200),
    is_active BOOLEAN,
    date_joined TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    longitude VARCHAR(100),
    latitude VARCHAR(100),
    device_id VARCHAR(100));

CREATE TABLE IF NOT EXISTS ResetRequests (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR (100) NOT NULL,
    token VARCHAR (100) NOT NULL,
    use_case VARCHAR (100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    consumed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    consumed BOOLEAN DEFAULT FALSE);
