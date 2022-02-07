import 'windi.css';
import 'floating-vue/dist/style.css';
import '~/compositions/useFavicon';

import { createPinia } from 'pinia';
import { createApp } from 'vue';

import App from '~/App.vue';
import FloatingVue from 'floating-vue'
import useEvents from '~/compositions/useEvents';
import { notifications } from '~/compositions/useNotifications';
import router from '~/router';

const app = createApp(App);

app.use(router);
app.use(notifications);
app.use(FloatingVue);
app.use(createPinia());
app.mount('#app');

useEvents();
