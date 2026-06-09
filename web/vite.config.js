import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';
import { readFileSync } from 'fs';

const packageJson = JSON.parse(readFileSync(new URL('./package.json', import.meta.url), 'utf8'));
const appVersion = process.env.VITE_APP_VERSION || readFileSync(new URL('../VERSION', import.meta.url), 'utf8').trim() || packageJson.version;
const buildTimestamp = process.env.VITE_BUILD_TIMESTAMP || new Date().toISOString();

export default defineConfig({
  plugins: [vue()],
  envPrefix: ['VITE_', 'VUE_APP_'],
  define: {
    __VUE_OPTIONS_API__: true,
    __VUE_PROD_DEVTOOLS__: false,
    __VUE_I18N_FULL_INSTALL__: true,
    __VUE_I18N_LEGACY_API__: true,
    __INTLIFY_PROD_DEVTOOLS__: false,
    __INTLIFY_JIT_COMPILATION__: false,
    __INTLIFY_DROP_MESSAGE_COMPILER__: false,
    'import.meta.env.VITE_APP_VERSION': JSON.stringify(appVersion),
    'import.meta.env.VITE_BUILD_TIMESTAMP': JSON.stringify(buildTimestamp),
    'process.env.VUE_APP_API_URL': JSON.stringify(process.env.VUE_APP_API_URL || ''),
    'process.env.VUE_APP_API_VERSION': JSON.stringify(process.env.VUE_APP_API_VERSION || '/api/v1')
  },
  resolve: {
    alias: [
      { find: '@', replacement: path.resolve(__dirname, './src') }
    ]
  },
  server: {
    port: 8080,
    proxy: {
      '/api': {
        target: 'http://localhost:22112',
        changeOrigin: true
      },
      '/swagger': {
        target: 'http://localhost:22112',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true
  },
  base: process.env.NODE_ENV === 'production' ? '/dist/' : '/'
});
