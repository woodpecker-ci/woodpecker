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
    path: '/repos/add',
    name: 'repo-add',
    component: (): Component => import('~/views/RepoAdd.vue'),
    meta: { authentication: 'required' },
  },
  {
    path: '/org/:ownerOrOrgId',
    component: (): Component => import('~/views/org/OrgWrapper.vue'),
    props: true,
    children: [
      {
        path: '',
        name: 'org',
        component: (): Component => import('~/views/org/OrgRepos.vue'),
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
    path: '/repos/:repoId',
    name: 'repo-wrapper',
    component: (): Component => import('~/views/repo/RepoWrapper.vue'),
    props: true,
    children: [
      {
        path: '',
        name: 'repo',
        component: (): Component => import('~/views/repo/RepoPipelines.vue'),
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
        path: 'pull-requests',
        name: 'repo-pull-requests',
        component: (): Component => import('~/views/repo/RepoPullRequests.vue'),
        meta: { repoHeader: true },
        props: (route) => ({ pullRequest: route.params.pullRequest }),
      },
      {
        path: 'pull-requests/:pullRequest',
        name: 'repo-pull-request',
        component: (): Component => import('~/views/repo/RepoPullRequest.vue'),
        meta: { repoHeader: true },
        props: (route) => ({ pullRequest: route.params.pullRequest }),
      },
      {
        path: 'pipeline/:pipelineId',
        component: (): Component => import('~/views/repo/pipeline/PipelineWrapper.vue'),
        props: true,
        children: [
          {
            path: ':stepId?',
            name: 'repo-pipeline',
            component: (): Component => import('~/views/repo/pipeline/Pipeline.vue'),
            props: true,
          },
          {
            path: 'changed-files',
            name: 'repo-pipeline-changed-files',
            component: (): Component => import('~/views/repo/pipeline/PipelineChangedFiles.vue'),
          },
          {
            path: 'config',
            name: 'repo-pipeline-config',
            component: (): Component => import('~/views/repo/pipeline/PipelineConfig.vue'),
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

  // TODO: deprecated routes => remove after some time
  {
    path: '/:ownerOrOrgId',
    redirect: (route) => ({ name: 'org', params: route.params }),
  },
  {
    path: '/:repoOwner/:repoName/:pathMatch(.*)*',
    redirect: (route) => {
      console.log(route);
      return { name: 'repo', params: route.params };
    },
  },
  {
    path: '/:repoOwner/:repoName',
    component: (): Component => import('~/views/repo/RepoWrapper.vue'),
    children: [
      // TODO: redirect to support backwards compatibility => remove after some time
      {
        path: ':pipelineId',
        redirect: (route) => ({ name: 'repo-pipeline', params: route.params }),
      },
      {
        path: 'build/:pipelineId',
        redirect: (route) => ({ name: 'repo-pipeline', params: route.params }),
        children: [
          {
            path: ':procId?',
            redirect: (route) => ({ name: 'repo-pipeline', params: route.params }),
          },
          {
            path: 'changed-files',
            redirect: (route) => ({ name: 'repo-pipeline-changed-files', params: route.params }),
          },
          {
            path: 'config',
            redirect: (route) => ({ name: 'repo-pipeline-config', params: route.params }),
          },
        ],
      },
    ],
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
