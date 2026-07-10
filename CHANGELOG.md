# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0.1] - 2026-07-11

### Added
- Enabled Row Level Security (RLS) on public database tables `projects`, `resume_config`, and `resume_waitlist`.
- Added public read-only access policies for `projects` and `resume_config`.
- Added an insert-only policy for `resume_waitlist` that restricts public access and validates email formats to protect user privacy.

## [0.1.0.0] - 2026-07-11

### Changed
- Migrated administrative authentication, file storage, and data models from Firebase to Supabase.
- Configured Content Security Policy (CSP) security headers in Go server middleware to allow connections to Supabase domains.
- Updated user session verification to decode and validate Supabase JWT tokens.

### Removed
- Deleted obsolete Firebase configuration and service initialization code.
