import type { RouteLocationNormalizedLoaded, Router } from 'vue-router';

import { usePipelineStore } from '~/store/pipelines';
import { useRepoStore } from '~/store/repos';

import useApiClient from './useApiClient';
import useConfig from './useConfig';

const apiClient = useApiClient();
let initialized = false;
let eventSource: EventSource | undefined;

function getAuthenticationMode(route: RouteLocationNormalizedLoaded) {
  return route.matched.toReversed().find((record) => record.meta.authentication != null)?.meta
    .authentication;
}

export function shouldSubscribeEvents(route: RouteLocationNormalizedLoaded): boolean {
  const authenticationMode = getAuthenticationMode(route);

  if (authenticationMode === 'guest-only') {
    return false;
  }

  if (authenticationMode === 'required' && !useConfig().user) {
    return false;
  }

  return true;
}

export function closeEvents() {
  eventSource?.close();
  eventSource = undefined;
  initialized = false;
}

function subscribeEvents() {
  if (initialized) {
    return;
  }

  const repoStore = useRepoStore();
  const pipelineStore = usePipelineStore();

  initialized = true;

  eventSource = apiClient.on((data) => {
    // contains repo update
    if (!data.repo) {
      return;
    }
    const { repo } = data;
    repoStore.setRepo(repo);

    // contains pipeline update
    if (!data.pipeline) {
      return;
    }
    const { pipeline } = data;
    pipelineStore.setPipeline(repo.id, pipeline);
  });
}

export function syncEventsSubscription(route: RouteLocationNormalizedLoaded) {
  if (shouldSubscribeEvents(route)) {
    subscribeEvents();
  } else {
    closeEvents();
  }
}

export function setupEvents(router: Router) {
  syncEventsSubscription(router.currentRoute.value);
  router.afterEach((to) => {
    syncEventsSubscription(to);
  });
}

export default subscribeEvents;
