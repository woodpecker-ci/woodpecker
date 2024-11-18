import type { Component } from 'vue';
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';

import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import useUserConfig from '~/compositions/useUserConfig';

const { rootPath } = useConfig();
const routes: RouteRecordRaw[] = [
  {
    path: `${rootPath}/`,
    name: 'home',
    redirect: `${rootPath}/repos`,
  },
  {
    path: `${rootPath}/repos`,
    component: (): Component => import('~/views/RouterView.vue'),
    children: [
      {
        path: '',
        name: 'repos',
        component: (): Component => import('~/views/Repos.vue'),
        meta: { authentication: 'required' },
      },
      {
        path: 'add',
        name: 'repo-add',
        component: (): Component => import('~/views/RepoAdd.vue'),
        meta: { authentication: 'required' },
      },
      {
        path: ':repoId',
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
            meta: { repoHeader: true },
            children: [
              {
                path: '',
                name: 'repo-branches',
                component: (): Component => import('~/views/repo/RepoBranches.vue'),
              },
              {
                path: ':branch',
                name: 'repo-branch',
                component: (): Component => import('~/views/repo/RepoBranch.vue'),
                props: (route) => ({ branch: route.params.branch }),
              },
            ],
          },

          {
            path: 'pull-requests',
            meta: { repoHeader: true },
            children: [
              {
                path: '',
                name: 'repo-pull-requests',
                component: (): Component => import('~/views/repo/RepoPullRequests.vue'),
              },
              {
                path: ':pullRequest',
                name: 'repo-pull-request',
                component: (): Component => import('~/views/repo/RepoPullRequest.vue'),
                props: (route) => ({ pullRequest: route.params.pullRequest }),
              },
            ],
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
              {
                path: 'errors',
                name: 'repo-pipeline-errors',
                component: (): Component => import('~/views/repo/pipeline/PipelineErrors.vue'),
                props: true,
              },
              {
                path: 'debug',
                name: 'repo-pipeline-debug',
                component: (): Component => import('~/views/repo/pipeline/PipelineDebug.vue'),
              },
            ],
          },
          {
            path: 'settings',
            component: (): Component => import('~/views/repo/settings/RepoSettings.vue'),
            meta: { authentication: 'required' },
            props: true,
            children: [
              {
                path: '',
                name: 'repo-settings',
                component: (): Component => import('~/views/repo/settings/General.vue'),
                props: true,
              },
              {
                path: 'secrets',
                name: 'repo-settings-secrets',
                component: (): Component => import('~/views/repo/settings/Secrets.vue'),
                props: true,
              },
              {
                path: 'registries',
                name: 'repo-settings-registries',
                component: (): Component => import('~/views/repo/settings/Registries.vue'),
                props: true,
              },
              {
                path: 'crons',
                name: 'repo-settings-crons',
                component: (): Component => import('~/views/repo/settings/Crons.vue'),
                props: true,
              },
              {
                path: 'badge',
                name: 'repo-settings-badge',
                component: (): Component => import('~/views/repo/settings/Badge.vue'),
                props: true,
              },
              {
                path: 'actions',
                name: 'repo-settings-actions',
                component: (): Component => import('~/views/repo/settings/Actions.vue'),
                props: true,
              },
            ],
          },
          {
            path: 'manual',
            name: 'repo-manual',
            component: (): Component => import('~/views/repo/RepoManualPipeline.vue'),
            meta: { authentication: 'required', repoHeader: true },
          },
        ],
      },
      {
        path: ':repoOwner/:repoName/:pathMatch(.*)*',
        component: (): Component => import('~/views/repo/RepoDeprecatedRedirect.vue'),
        props: true,
      },
    ],
  },
  {
    path: `${rootPath}/orgs/:orgId`,
    component: (): Component => import('~/views/org/OrgWrapper.vue'),
    props: true,
    children: [
      {
        path: '',
        name: 'org',
        component: (): Component => import('~/views/org/OrgRepos.vue'),
        props: true,
      },
      {
        path: 'settings',
        name: 'org-settings',
        component: (): Component => import('~/views/org/settings/OrgSettingsWrapper.vue'),
        meta: { authentication: 'required' },
        props: true,
        children: [
          {
            path: 'secrets',
            name: 'org-settings-secrets',
            component: (): Component => import('~/views/org/settings/OrgSecrets.vue'),
            props: true,
          },
          {
            path: 'registries',
            name: 'org-settings-registries',
            component: (): Component => import('~/views/org/settings/OrgRegistries.vue'),
            props: true,
          },
          {
            path: 'agents',
            name: 'org-settings-agents',
            component: (): Component => import('~/views/org/settings/OrgAgents.vue'),
            props: true,
          },
        ],
      },
    ],
  },
  {
    path: `${rootPath}/org/:orgName/:pathMatch(.*)*`,
    component: (): Component => import('~/views/org/OrgDeprecatedRedirect.vue'),
    props: true,
  },
  {
    path: `${rootPath}/admin`,
    component: (): Component => import('~/views/admin/AdminSettingsWrapper.vue'),
    props: true,
    meta: { authentication: 'required' },
    children: [
      {
        path: '',
        name: 'admin-settings',
        component: (): Component => import('~/views/admin/AdminInfo.vue'),
        props: true,
      },
      {
        path: 'secrets',
        name: 'admin-settings-secrets',
        component: (): Component => import('~/views/admin/AdminSecrets.vue'),
        props: true,
      },
      {
        path: 'registries',
        name: 'admin-settings-registries',
        component: (): Component => import('~/views/admin/AdminRegistries.vue'),
        props: true,
      },
      {
        path: 'repos',
        name: 'admin-settings-repos',
        component: (): Component => import('~/views/admin/AdminRepos.vue'),
        props: true,
      },
      {
        path: 'users',
        name: 'admin-settings-users',
        component: (): Component => import('~/views/admin/AdminUsers.vue'),
        props: true,
      },
      {
        path: 'orgs',
        name: 'admin-settings-orgs',
        component: (): Component => import('~/views/admin/AdminOrgs.vue'),
        props: true,
      },
      {
        path: 'agents',
        name: 'admin-settings-agents',
        component: (): Component => import('~/views/admin/AdminAgents.vue'),
        props: true,
      },
      {
        path: 'queue',
        name: 'admin-settings-queue',
        component: (): Component => import('~/views/admin/AdminQueue.vue'),
        props: true,
      },
    ],
  },

  {
    path: `${rootPath}/user`,
    component: (): Component => import('~/views/user/UserWrapper.vue'),
    meta: { authentication: 'required' },
    props: true,
    children: [
      {
        path: '',
        name: 'user',
        component: (): Component => import('~/views/user/UserGeneral.vue'),
        props: true,
      },
      {
        path: 'secrets',
        name: 'user-secrets',
        component: (): Component => import('~/views/user/UserSecrets.vue'),
        props: true,
      },
      {
        path: 'registries',
        name: 'user-registries',
        component: (): Component => import('~/views/user/UserRegistries.vue'),
        props: true,
      },
      {
        path: 'cli-and-api',
        name: 'user-cli-and-api',
        component: (): Component => import('~/views/user/UserCLIAndAPI.vue'),
        props: true,
      },
      {
        path: 'agents',
        name: 'user-agents',
        component: (): Component => import('~/views/user/UserAgents.vue'),
        props: true,
      },
    ],
  },
  {
    path: `${rootPath}/login`,
    name: 'login',
    component: (): Component => import('~/views/Login.vue'),
    meta: { blank: true },
    props: true,
  },
  {
    path: `${rootPath}/cli/auth`,
    component: (): Component => import('~/views/cli/Auth.vue'),
    meta: { authentication: 'required' },
  },

  // TODO: deprecated routes => remove after some time
  {
    path: `${rootPath}/:ownerOrOrgId`,
    redirect: (route) => ({ name: 'org', params: route.params }),
  },
  {
    path: `${rootPath}/:repoOwner/:repoName/:pathMatch(.*)*`,
    component: (): Component => import('~/views/repo/RepoDeprecatedRedirect.vue'),
    props: true,
  },

  // not found handler
  {
    path: `${rootPath}/:pathMatch(.*)*`,
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
  if (redirectUrl !== '' && to.name !== 'login') {
    config.setUserConfig('redirectUrl', '');
    next(redirectUrl);
  }

  const authentication = useAuthentication();
  const authenticationRequired = to.matched.some((record) => record.meta.authentication === 'required');
  if (authenticationRequired && !authentication.isAuthenticated) {
    next({ name: 'login', query: { url: to.fullPath } });
    return;
  }

  next();
});

export default router;
