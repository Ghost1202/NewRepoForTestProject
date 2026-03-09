BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY, 
    user_login VARCHAR(100) UNIQUE, 
    email VARCHAR(255) UNIQUE,       
    phone VARCHAR(20) UNIQUE,        
    pass_hash BYTEA, 
    google_id VARCHAR(255) UNIQUE, 
    role VARCHAR(50) NOT NULL DEFAULT 'user'
);

CREATE TABLE IF NOT EXISTS user_info(
    id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, 
    first_name VARCHAR(100),    
    last_name VARCHAR(100),  
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMIT;