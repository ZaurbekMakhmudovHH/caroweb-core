CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password VARCHAR(255) NOT NULL,
                       role VARCHAR(20) NOT NULL DEFAULT 'ROLE_HOMEOWNER',
                       email_confirmation_token VARCHAR(64),
                       last_confirmation_sent_at TIMESTAMP,
                       email_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE UNIQUE INDEX idx_users_email_unique ON users (email);

CREATE TABLE user_profiles (
                               user_id UUID PRIMARY KEY,
                               salutation VARCHAR(20) NOT NULL,
                               title VARCHAR(50),
                               first_name VARCHAR(100) NOT NULL,
                               last_name VARCHAR(100) NOT NULL,
                               street VARCHAR(100) NOT NULL,
                               house_number VARCHAR(10) NOT NULL,
                               postal_code VARCHAR(10) NOT NULL,
                               city VARCHAR(100) NOT NULL,
                               is_verified BOOLEAN NOT NULL DEFAULT FALSE,
                               updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               CONSTRAINT fk_user
                                   FOREIGN KEY(user_id)
                                       REFERENCES users(id)
                                       ON DELETE CASCADE
);

CREATE TABLE refresh_tokens (
                                id UUID PRIMARY KEY,
                                user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                token TEXT NOT NULL UNIQUE,
                                expires_at TIMESTAMP NOT NULL,
                                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);