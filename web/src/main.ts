import '~/compositions/useFavicon';
import '~/tailwind.css'
import '~/style.css';

import { createPinia } from 'pinia';
import { createApp } from 'vue';

import App from '~/App.vue';
import useEvents from '~/compositions/useEvents';
import { i18n } from '~/compositions/useI18n';
import { notifications } from '~/compositions/useNotifications';
import router from '~/router';

// eslint-disable-next-line ts/no-unsafe-argument
const app = createApp(App);

app.use(router);
app.use(notifications);
app.use(i18n);

app.use(createPinia());
app.mount('#app');

useEvents();
