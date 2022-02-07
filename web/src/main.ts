import 'windi.css';
import 'floating-vue/dist/style.css'; // eslint-disable-line no-restricted-imports
import '~/compositions/useFavicon';

import FloatingVue from 'floating-vue';
import { createPinia } from 'pinia';
import { createApp } from 'vue';

import App from '~/App.vue';
import useEvents from '~/compositions/useEvents';
import { notifications } from '~/compositions/useNotifications';
import router from '~/router';

const app = createApp(App);

app.use(router);
app.use(notifications);
app.use(FloatingVue);
app.directive('tooltip', FloatingVue.VTooltip);
app.use(createPinia());
app.mount('#app');

useEvents();
