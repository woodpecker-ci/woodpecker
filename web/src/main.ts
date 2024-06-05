import 'windi.css';
import '~/compositions/useFavicon';
import '~/style.css';
import { createPinia } from 'pinia';
import type { Component, ComputedOptions, MethodOptions } from 'vue';
import { createApp } from 'vue';
import App from '~/App.vue';
import useEvents from '~/compositions/useEvents';
import { i18n } from '~/compositions/useI18n';
import { notifications } from '~/compositions/useNotifications';
import router from '~/router';

const app = createApp(App as Component<any, any, any, ComputedOptions, MethodOptions, any, any>);

app.use(router);
app.use(notifications);
app.use(i18n);

app.use(createPinia());
app.mount('#app');

useEvents();
