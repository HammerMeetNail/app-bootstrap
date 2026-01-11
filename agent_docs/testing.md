# Testing Guide

## Local development
- `make local` to start the stack (Podman-first).
- `make logs` to follow container logs.
- `make down` to stop containers.

## Tests
- `make test` runs Go + JS tests in a container.
- `make e2e` runs Playwright tests in a fresh stack.

## JS tests
- Node runner: `node web/static/js/tests/runner.js`
- Browser runner: `web/static/js/tests/runner.html`
