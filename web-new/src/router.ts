import { Component } from 'vue';
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: (): Component => import('~/views/Home.vue'),
  },
  {
    path: '/projects',
    name: 'projects',
    component: (): Component => import('~/views/Home.vue'),
    meta: { authentication: 'required' },
  },
  {
    path: '/do-login/:origin?',
    name: 'login',
    component: (): Component => import('~/views/Login.vue'),
    props: true,
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: (): Component => import('~/views/NotFound.vue'),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to, _, next) => {
  if (to.meta.authentication === 'required') {
    next({ name: 'login', params: { origin: to.fullPath } });
    return;
  }

  next();
});

export default router;
