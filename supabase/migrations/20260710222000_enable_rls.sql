-- Enable Row Level Security (RLS) on tables exposed to PostgREST
ALTER TABLE public.projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.resume_config ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.resume_waitlist ENABLE ROW LEVEL SECURITY;

-- 1. Projects Table Policy
-- Projects are public portfolio data, so anyone should be allowed to view (SELECT) them.
CREATE POLICY "Allow public read access to projects"
ON public.projects
FOR SELECT
TO public
USING (true);

-- 2. Resume Config Table Policy
-- Resume configuration is public portfolio metadata, so anyone should be allowed to view (SELECT) it.
CREATE POLICY "Allow public read access to resume_config"
ON public.resume_config
FOR SELECT
TO public
USING (true);

-- 3. Resume Waitlist Table Policy
-- Any visitor to the website can submit (INSERT) their email to the waitlist.
-- We require that the email is not null and contains an '@' character to prevent empty/malformed entries.
CREATE POLICY "Allow public insert access to resume_waitlist"
ON public.resume_waitlist
FOR INSERT
TO public
WITH CHECK (email IS NOT NULL AND position('@' in email) > 0);
