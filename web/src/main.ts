import 'windi.css';

import { createApp } from 'vue';
import { createPinia } from 'pinia';

import App from '~/App.vue';
import router from '~/router';
import { notifications } from '~/compositions/useNotifications';
import useEvents from '~/compositions/useEvents';

const app = createApp(App);

app.use(router);
app.use(notifications);
app.use(createPinia());
app.mount('#app');

useEvents();
