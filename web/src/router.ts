import type { Component } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import useUserConfig from '~/compositions/useUserConfig';

declare module 'vue-router' {
  interface RouteMeta {
    authentication: 'required' | 'optional';
    repoHeader?: true;
    layout?: 'default' | 'blank';
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    redirect: { name: 'repos' },
  },
  {
    path: '/repos',
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
        component: (): Component => import('~/views/repo/RepoWrapper.vue'),
        props: true,
        meta: { authentication: 'optional' },
        children: [
          {
            path: '',
            name: 'repo',
            component: (): Component => import('~/views/repo/RepoPipelines.vue'),
            meta: { repoHeader: true, authentication: 'optional' },
          },
          {
            path: 'branches',
            meta: { repoHeader: true, authentication: 'optional' },
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
            meta: { repoHeader: true, authentication: 'optional' },
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
            meta: { authentication: 'optional' },
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
                props: true,
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
        meta: { authentication: 'optional' },
      },
    ],
  },
  {
    path: '/orgs/:orgId',
    component: (): Component => import('~/views/org/OrgWrapper.vue'),
    props: true,
    children: [
      {
        path: '',
        name: 'org',
        component: (): Component => import('~/views/org/OrgRepos.vue'),
        props: true,
        meta: { authentication: 'optional' },
      },
      {
        path: 'settings',
        component: (): Component => import('~/views/org/settings/OrgSettingsWrapper.vue'),
        meta: { authentication: 'required' },
        props: true,
        children: [
          {
            path: '',
            name: 'org-settings',
            redirect: { name: 'org-settings-secrets' },
          },
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
    path: '/org/:orgName/:pathMatch(.*)*',
    component: (): Component => import('~/views/org/OrgDeprecatedRedirect.vue'),
    props: true,
    meta: { authentication: 'optional' },
  },
  {
    path: '/admin',
    component: (): Component => import('~/views/admin/AdminSettingsWrapper.vue'),
    meta: { authentication: 'required' },
    children: [
      {
        path: '',
        name: 'admin-settings',
        component: (): Component => import('~/views/admin/AdminInfo.vue'),
      },
      {
        path: 'secrets',
        name: 'admin-settings-secrets',
        component: (): Component => import('~/views/admin/AdminSecrets.vue'),
      },
      {
        path: 'registries',
        name: 'admin-settings-registries',
        component: (): Component => import('~/views/admin/AdminRegistries.vue'),
      },
      {
        path: 'repos',
        name: 'admin-settings-repos',
        component: (): Component => import('~/views/admin/AdminRepos.vue'),
      },
      {
        path: 'users',
        name: 'admin-settings-users',
        component: (): Component => import('~/views/admin/AdminUsers.vue'),
      },
      {
        path: 'orgs',
        name: 'admin-settings-orgs',
        component: (): Component => import('~/views/admin/AdminOrgs.vue'),
      },
      {
        path: 'agents',
        name: 'admin-settings-agents',
        component: (): Component => import('~/views/admin/AdminAgents.vue'),
      },
      {
        path: 'queue',
        name: 'admin-settings-queue',
        component: (): Component => import('~/views/admin/AdminQueue.vue'),
      },
      {
        path: 'forges',
        component: (): Component => import('~/components/layout/RouteWrapper.vue'),
        props: true,
        children: [
          {
            path: '',
            name: 'admin-settings-forges',
            component: (): Component => import('~/views/admin/forges/AdminForges.vue'),
          },
          {
            path: ':forgeId',
            name: 'admin-settings-forge',
            component: (): Component => import('~/views/admin/forges/AdminForge.vue'),
            props: true,
          },
          {
            path: 'create',
            name: 'admin-settings-forge-create',
            component: (): Component => import('~/views/admin/forges/AdminForgeCreate.vue'),
          },
        ],
      },
    ],
  },

  {
    path: '/user',
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
    path: '/login',
    name: 'login',
    component: (): Component => import('~/views/Login.vue'),
    meta: { layout: 'blank', authentication: 'optional' },
  },
  {
    path: '/cli/auth',
    component: (): Component => import('~/views/cli/Auth.vue'),
    meta: { authentication: 'required' },
  },

  // TODO: deprecated routes => remove after some time
  {
    path: '/:ownerOrOrgId',
    redirect: (route) => ({ name: 'org', params: route.params }),
    meta: { authentication: 'optional' },
  },
  {
    path: '/:repoOwner/:repoName/:pathMatch(.*)*',
    component: (): Component => import('~/views/repo/RepoDeprecatedRedirect.vue'),
    props: true,
    meta: { authentication: 'optional' },
  },

  // not found handler
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: (): Component => import('~/views/NotFound.vue'),
    meta: { authentication: 'optional' },
  },
];

const { rootPath } = useConfig();
const router = createRouter({
  history: createWebHistory(),
  routes: routes.map((r) => ({ ...r, path: `${rootPath}${r.path}` })),
});

router.beforeEach(async (to, _, next) => {
  const config = useUserConfig();
  const { redirectUrl } = config.userConfig.value;
  if (redirectUrl !== '' && to.name !== 'login') {
    config.setUserConfig('redirectUrl', '');
    next(redirectUrl);
  }

  const authentication = useAuthentication();
  const authenticationMode =
    to.matched.find((record) => record.meta.authentication === 'required')?.meta.authentication ?? 'required';
  if (authenticationMode === 'required' && !authentication.isAuthenticated) {
    next({ name: 'login', query: { url: to.fullPath } });
    return;
  }

  next();
});

export default router;
