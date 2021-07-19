import { Component } from 'vue';
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import { isAuthenticated } from './compositions/useAuthentication';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: (): Component => import('~/views/Home.vue'),
  },
  {
    path: '/repos',
    name: 'repos',
    component: (): Component => import('~/views/repos/Repos.vue'),
    meta: { authentication: 'required' },
  },
  {
    path: '/repo/add',
    name: 'repo-add',
    component: (): Component => import('~/views/repos/RepoAdd.vue'),
    meta: { authentication: 'required' },
  },
  {
    path: '/repo/:repoOwner/:repoId',
    name: 'repo',
    component: (): Component => import('~/views/repos/Repo.vue'),
    meta: { authentication: 'required' },
    props: true,
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
  if (to.meta.authentication === 'required' && !isAuthenticated()) {
    next({ name: 'login', params: { origin: to.fullPath } });
    return;
  }

  next();
});

export default router;
