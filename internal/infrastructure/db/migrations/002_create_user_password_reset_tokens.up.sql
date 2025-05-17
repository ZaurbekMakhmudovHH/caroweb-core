CREATE TABLE user_password_reset_tokens (
                                       id UUID PRIMARY KEY,
                                       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                       token TEXT NOT NULL,
                                       created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                       used_at TIMESTAMP WITH TIME ZONE,
                                       expires_at TIMESTAMP WITH TIME ZONE,
                                       UNIQUE(token)
);