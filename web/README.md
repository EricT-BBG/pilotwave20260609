# Pilotwave Web

The frontend is a Vue 3 SPA built with Vite. It uses Vue Router 4, Vuex 4,
Vue I18n, PrimeVue, and Vitest.

## Setup

```shell
npm ci --include=dev
```

## Development

```shell
npm run dev
```

Point the app at a local Pilotwave API with Vite env vars when needed:

```shell
VITE_API_URL=http://127.0.0.1:22112 VITE_API_VERSION=v1 npm run dev
```

The runtime still accepts legacy `VUE_APP_API_URL` and `VUE_APP_API_VERSION`
names during the migration.

## Verification

```shell
npm run lint
npm test
npm run build
```

The repository-level `make verify` command runs the frontend lint, tests, build,
static asset generation, and backend build.
