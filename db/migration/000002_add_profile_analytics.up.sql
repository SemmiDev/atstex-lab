CREATE TABLE IF NOT EXISTS profile_analytics (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES cv_profiles(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referer TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_profile_analytics_user_id ON profile_analytics(user_id);
CREATE INDEX IF NOT EXISTS idx_profile_analytics_profile_id ON profile_analytics(profile_id);
