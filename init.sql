CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    google_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    picture TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    ai_tokens_used BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_users_google_id ON users(google_id);

CREATE TABLE IF NOT EXISTS cv_profiles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    biodata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_cv_profiles_user_id ON cv_profiles(user_id);

-- Migration helper: add columns if they don't exist (safe to re-run)
DO $$ BEGIN
    ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';
    ALTER TABLE users ADD COLUMN IF NOT EXISTS ai_tokens_used BIGINT NOT NULL DEFAULT 0;
    ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(50) UNIQUE;
    ALTER TABLE users ADD COLUMN IF NOT EXISTS is_blocked BOOLEAN NOT NULL DEFAULT false;
    ALTER TABLE cv_profiles ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT false;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

CREATE TABLE IF NOT EXISTS feedbacks (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    admin_reply TEXT,
    replied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_feedbacks_user_id ON feedbacks(user_id);

CREATE TABLE IF NOT EXISTS cv_reviews (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES cv_profiles(id) ON DELETE CASCADE,
    profile_title VARCHAR(255) NOT NULL DEFAULT '',
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    score INTEGER NOT NULL DEFAULT 0,
    strengths TEXT NOT NULL DEFAULT '',
    improvements TEXT NOT NULL DEFAULT '',
    recommendations TEXT NOT NULL DEFAULT '',
    tokens_used BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cv_reviews_user_id ON cv_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_cv_reviews_profile_id ON cv_reviews(profile_id);

CREATE TABLE IF NOT EXISTS cover_letters (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES cv_profiles(id) ON DELETE CASCADE,
    profile_title VARCHAR(255) NOT NULL DEFAULT '',
    job_description TEXT NOT NULL DEFAULT '',
    cover_letter_text TEXT NOT NULL DEFAULT '',
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    tokens_used BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_cover_letters_user_id ON cover_letters(user_id);
CREATE INDEX IF NOT EXISTS idx_cover_letters_profile_id ON cover_letters(profile_id);

CREATE TABLE IF NOT EXISTS job_applications (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cv_profile_id UUID NOT NULL REFERENCES cv_profiles(id) ON DELETE CASCADE,
    company VARCHAR(255) NOT NULL,
    job_title VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Applied',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_job_applications_user_id ON job_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_cv_profile_id ON job_applications(cv_profile_id);

-- Migration: add deadline column to job_applications
DO $$ BEGIN
    ALTER TABLE job_applications ADD COLUMN IF NOT EXISTS deadline DATE;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS ats_simulations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES cv_profiles(id) ON DELETE CASCADE,
    job_description TEXT NOT NULL DEFAULT '',
    score INTEGER NOT NULL DEFAULT 0,
    missing_keywords JSONB NOT NULL DEFAULT '[]',
    recommendations TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ats_simulations_user_id ON ats_simulations(user_id);
CREATE INDEX IF NOT EXISTS idx_ats_simulations_profile_id ON ats_simulations(profile_id);

CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    price_idr BIGINT NOT NULL,
    duration_months INTEGER NOT NULL DEFAULT 1,
    max_cv_profiles INTEGER NOT NULL, -- -1 for unlimited
    max_cv_reviews INTEGER NOT NULL, -- -1 for unlimited
    max_ats_simulations INTEGER NOT NULL,
    max_cover_letters INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions(user_id);

-- Default Subscription Plans
INSERT INTO subscription_plans (id, name, price_idr, duration_months, max_cv_profiles, max_cv_reviews, max_ats_simulations, max_cover_letters, is_active)
SELECT
    '018e9c40-1111-7000-8000-000000000001'::uuid, 'Gratis', 0, 1, 1, 5, 5, 10, true
WHERE NOT EXISTS (SELECT 1 FROM subscription_plans WHERE name = 'Gratis');

INSERT INTO subscription_plans (id, name, price_idr, duration_months, max_cv_profiles, max_cv_reviews, max_ats_simulations, max_cover_letters, is_active)
SELECT
    '018e9c40-2222-7000-8000-000000000002'::uuid, 'Basic', 20000, 1, 3, 10, 10, 10, true
WHERE NOT EXISTS (SELECT 1 FROM subscription_plans WHERE name = 'Basic');

INSERT INTO subscription_plans (id, name, price_idr, duration_months, max_cv_profiles, max_cv_reviews, max_ats_simulations, max_cover_letters, is_active)
SELECT
    '018e9c40-3333-7000-8000-000000000003'::uuid, 'Pro', 30000, 1, -1, 50, 50, 50, true
WHERE NOT EXISTS (SELECT 1 FROM subscription_plans WHERE name = 'Pro');

CREATE TABLE IF NOT EXISTS interview_preps (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES cv_profiles(id) ON DELETE CASCADE,
    job_description TEXT NOT NULL DEFAULT '',
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    questions JSONB NOT NULL DEFAULT '[]',
    tokens_used BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_interview_preps_user_id ON interview_preps(user_id);
CREATE INDEX IF NOT EXISTS idx_interview_preps_profile_id ON interview_preps(profile_id);
