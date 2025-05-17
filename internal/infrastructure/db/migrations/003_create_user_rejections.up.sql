CREATE TABLE user_rejections (
                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                 errors JSONB,
                                 rejected_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
