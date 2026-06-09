import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';
import i18n from './plugins/i18n';
import './styles/app.css';
import 'material-design-icons-iconfont/dist/material-design-icons.css';
import '@mdi/font/css/materialdesignicons.css';
import './lib/http';

const app = createApp(App);

app.use(store);
app.use(router);
app.use(i18n);

app.mount('#app');
