-- Create projects table
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    subtitle TEXT,
    description TEXT,
    hero_gif TEXT,
    challenge TEXT,
    solution TEXT,
    architecture TEXT,
    arch_diagram TEXT,
    internal_flow TEXT[] DEFAULT '{}',
    tech_stack TEXT[] DEFAULT '{}',
    display_stack TEXT[] DEFAULT '{}',
    key_features TEXT[] DEFAULT '{}',
    live_url TEXT,
    github_url TEXT,
    metrics JSONB DEFAULT '{}'::jsonb,
    deployment TEXT
);

-- Create resume config table
CREATE TABLE IF NOT EXISTS resume_config (
    id TEXT PRIMARY KEY DEFAULT 'default',
    is_coming_soon BOOLEAN NOT NULL DEFAULT true,
    pdf_storage_uri TEXT
);

-- Insert default resume config row
INSERT INTO resume_config (id, is_coming_soon, pdf_storage_uri)
VALUES ('default', true, '')
ON CONFLICT (id) DO NOTHING;

-- Create resume waitlist table
CREATE TABLE IF NOT EXISTS resume_waitlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Create storage bucket for resumes
INSERT INTO storage.buckets (id, name, public) 
VALUES ('resumes', 'resumes', false)
ON CONFLICT (id) DO NOTHING;
