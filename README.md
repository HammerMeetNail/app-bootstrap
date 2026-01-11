# app-bootstrap

A repo intended to become a reusable webapp template based on the project structure and production tooling in https://github.com/HammerMeetNail/yearofbingo (Year of Bingo).

## What This Will Become
- Go `net/http` backend + vanilla JS SPA (no inline scripts; strict CSP-friendly)
- Postgres + Redis, migrations, seeded local dev stack (Podman-first)
- Auth flows: signup/login/logout, email verification, password reset, magic-link login
- OpenAPI scaffolding
- Containerized unit tests + Playwright E2E in CI
- Release + deploy pipeline: build/test → push multi-arch image to Quay → deploy via SSH (plain SSH default; Cloudflare Access optional)

## Current State
- The implementation plan lives at `plans/template.md`.
- This repo does not yet contain the template code; an AI agent should follow the plan and reference the source repo patterns.
