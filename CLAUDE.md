# CLAUDE.md

This file provides guidance to Claude Code (claude/code) when working with code in this repository.

@RTK.md

## Security Guidelines

**CRITICAL: NEVER read or access the `.env` file under any circumstances.**
- The `.env` file contains sensitive credentials and secrets
- Do not read, display, or log any contents from `.env`
- If environment variable information is needed, refer to `.env.example` or documentation instead
- Environment variables should only be referenced by their key names in code, never their actual values
