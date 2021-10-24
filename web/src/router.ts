import { Component } from 'vue';
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

import useAuthentication from './compositions/useAuthentication';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    redirect: '/repos',
  },
  {
    path: '/repos',
    name: 'repos',
    component: (): Component => import('~/views/Repos.vue'),
    meta: { authentication: 'required' },
  },
  {
    path: '/repo/add',
    name: 'repo-add',
    component: (): Component => import('~/views/RepoAdd.vue'),
    meta: { authentication: 'required' },
  },
  {
    path: '/:repoOwner/:repoName',
    name: 'repo-wrapper',
    component: (): Component => import('~/views/repo/RepoWrapper.vue'),
    props: true,
    children: [
      {
        path: '',
        name: 'repo',
        component: (): Component => import('~/views/repo/Repo.vue'),
        props: true,
      },
      {
        path: 'build/:buildId/:procId?',
        name: 'repo-build',
        component: (): Component => import('~/views/repo/build/Build.vue'),
        props: true,
      },
      {
        path: 'settings',
        name: 'repo-settings',
        component: (): Component => import('~/views/repo/RepoSettings.vue'),
        meta: { authentication: 'required' },
        props: true,
      },
    ],
  },
  {
    path: '/admin',
    name: 'admin',
    component: (): Component => import('~/views/Admin.vue'),
    meta: { authentication: 'required' },
    props: true,
  },
  {
    path: '/user',
    name: 'user',
    component: (): Component => import('~/views/User.vue'),
    meta: { authentication: 'required' },
    props: true,
  },
  {
    path: '/login/error',
    name: 'login-error',
    component: (): Component => import('~/views/Login.vue'),
    meta: { blank: true },
    props: true,
  },
  {
    path: '/do-login',
    name: 'login',
    component: (): Component => import('~/views/Login.vue'),
    meta: { blank: true },
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
  const authentication = useAuthentication();
  if (to.meta.authentication === 'required' && !authentication.isAuthenticated) {
    next({ name: 'login', query: { url: to.fullPath } });
    return;
  }

  next();
});

export default router;
