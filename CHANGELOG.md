# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to Semantic Versioning.

## [0.1.0.0] - 2026-07-11

### Changed
- Migrated administrative authentication, file storage, and data models from Firebase to Supabase.
- Configured Content Security Policy (CSP) security headers in Go server middleware to allow connections to Supabase domains.
- Updated user session verification to decode and validate Supabase JWT tokens.

### Removed
- Deleted obsolete Firebase configuration and service initialization code.
