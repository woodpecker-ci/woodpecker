import 'windi.css';

import { createApp } from 'vue';

import App from '~/App.vue';
import router from '~/router';
import { notifications } from '~/compositions/useNotifications';

import TimeAgo from 'javascript-time-ago';
import en from 'javascript-time-ago/locale/en';
TimeAgo.addDefaultLocale(en);

const app = createApp(App);

app.use(router);
app.use(notifications);
app.mount('#app');
