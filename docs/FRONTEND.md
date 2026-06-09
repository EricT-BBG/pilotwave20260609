# Frontend Notes

Pilotwave's active frontend is the Vue 3 SPA in `web/`.

## Current Stack

- Vue 3.5
- Vite 6
- Vue Router 4
- Vuex 4
- Vue I18n 11
- PrimeVue 4 plus native markup and shared CSS in `web/src/styles/app.css`
- Vitest for unit tests
- ESLint 9 with `eslint-plugin-vue`

The old Vue 2, Vue CLI, Vuetify, and Vuelidate runtime paths have been removed. Do not reintroduce compatibility bridge files, Vuetify templates, `validationMixin`, or `$v` form state.

## Structure

- `web/src/views/`: page-level screens
- `web/src/components/`: reusable UI blocks
- `web/src/router/`: route table and guards
- `web/src/store/`: Vuex modules and API actions
- `web/src/plugins/translate/`: visible UI copy
- `web/src/lib/`: shared runtime helpers
- `web/src/styles/app.css`: shared application styling

The application is still page-driven and Vuex-driven. API calls remain concentrated in Vuex action modules and shared helpers; keep request and response shapes stable unless the backend contract is intentionally changing.

## Istio Management UX

Visible UI copy should use Istio resource names. The frontend still has legacy
`router` file, route, and Vuex module names, but customer-facing labels should
say `VirtualService`, not Router.

Gateway, VirtualService, RequestAuthentication, and AuthorizationPolicy detail
pages are read-only on first load. The detail toolbar exposes edit actions; edit
mode should be scoped to the selected area:

- Gateway basic settings and selector labels open through the detail edit flow.
- Gateway and VirtualService association tabs have explicit association edit
  mode and should show pending changes such as currently enabled, added now,
  and unchecked now.
- VirtualService route rules are read-only until the route-rule edit action is
  used. Child route form components should receive and enforce a `readonly`
  prop so store mutations cannot fire from disabled views.
- API authentication and black/white list pages use the same inspect-first
  detail layout before loading their edit forms.

Namespace selectors should expose Istio sidecar injection status. When creating
cluster-backed Istio resources in a namespace without injection labels, warn the
user before submitting.

Gateway selector labels are edited as Kubernetes label key/value pairs. The
default `istio=ingressgateway` selector is only a fallback when the user leaves
selector labels empty; user-provided labels must be preserved.

## Development

```sh
cd web
npm ci --include=dev
npm run dev
```

The Vite dev server proxies `/api` and `/swagger` to `http://localhost:22112`.

Runtime environment variables should use `VITE_*`. The Vite config still accepts `VUE_APP_API_URL` and `VUE_APP_API_VERSION` as compatibility names for existing local scripts.

## Verification

Use focused checks while editing the frontend:

```sh
cd web && npm run lint
cd web && npm test
cd web && npm run build
```

Use repo-level checks before larger handoffs:

```sh
make verify
```

For regression scans after touching shared pages or form components:

```sh
rtk rg -n "<v-|</v-|validationMixin|vuelidate|\\$v|legacyVuetify|legacyVuelidate" web/src web/package.json web/vite.config.js --glob '!**/__tests__/**'
```

The scan should return no runtime usage.
