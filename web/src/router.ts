import { Component } from 'vue';
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

import useAuthentication from '~/compositions/useAuthentication';
import useUserConfig from '~/compositions/useUserConfig';

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
    path: '/:repoOwner',
    name: 'repos-owner',
    component: (): Component => import('~/views/ReposOwner.vue'),
    props: true,
  },
  {
    path: '/org/:repoOwner',
    component: (): Component => import('~/views/org/OrgWrapper.vue'),
    props: true,
    children: [
      {
        path: '',
        name: 'org',
        redirect: (route) => ({ name: 'repos-owner', params: route.params }),
      },
      {
        path: 'settings',
        name: 'org-settings',
        component: (): Component => import('~/views/org/OrgSettings.vue'),
        meta: { authentication: 'required' },
        props: true,
      },
    ],
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
        component: (): Component => import('~/views/repo/RepoBuilds.vue'),
        meta: { repoHeader: true },
      },
      {
        path: 'branches',
        name: 'repo-branches',
        component: (): Component => import('~/views/repo/RepoBranches.vue'),
        meta: { repoHeader: true },
        props: (route) => ({ branch: route.params.branch }),
      },
      {
        path: 'branches/:branch',
        name: 'repo-branch',
        component: (): Component => import('~/views/repo/RepoBranch.vue'),
        meta: { repoHeader: true },
        props: (route) => ({ branch: route.params.branch }),
      },
      {
        path: 'build/:buildId',
        component: (): Component => import('~/views/repo/build/BuildWrapper.vue'),
        props: true,
        children: [
          {
            path: ':procId?',
            name: 'repo-build',
            component: (): Component => import('~/views/repo/build/Build.vue'),
            props: true,
          },
          {
            path: 'changed-files',
            name: 'repo-build-changed-files',
            component: (): Component => import('~/views/repo/build/BuildChangedFiles.vue'),
          },
          {
            path: 'config',
            name: 'repo-build-config',
            component: (): Component => import('~/views/repo/build/BuildConfig.vue'),
            props: true,
          },
        ],
      },
      {
        path: 'settings',
        name: 'repo-settings',
        component: (): Component => import('~/views/repo/RepoSettings.vue'),
        meta: { authentication: 'required' },
        props: true,
      },
      // TODO: redirect to support backwards compatibility => remove after some time
      {
        path: ':buildId',
        redirect: (route) => ({ name: 'repo-build', params: route.params }),
      },
    ],
  },
  {
    path: '/admin',
    name: 'admin',
    component: (): Component => import('~/views/admin/Admin.vue'),
    meta: { authentication: 'required' },
    props: true,
  },
  {
    path: '/admin/settings',
    name: 'admin-settings',
    component: (): Component => import('~/views/admin/AdminSettings.vue'),
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
  const config = useUserConfig();
  const { redirectUrl } = config.userConfig.value;
  if (redirectUrl !== '') {
    config.setUserConfig('redirectUrl', '');
    next(redirectUrl);
  }

  const authentication = useAuthentication();
  if (to.meta.authentication === 'required' && !authentication.isAuthenticated) {
    next({ name: 'login', query: { url: to.fullPath } });
    return;
  }

  next();
});

export default router;
