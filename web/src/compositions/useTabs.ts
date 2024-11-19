import { ref } from 'vue';
import type { RouteLocationRaw } from 'vue-router';

import type { IconNames } from '~/components/atomic/Icon.vue';

import { inject, provide } from './useInjectProvide';

export interface Tab {
  to: RouteLocationRaw;
  title: string;
  icon?: IconNames;
  iconClass?: string;
  matchChildren?: boolean;
}

export function useTabsProvider() {
  const tabs = ref<Tab[]>([]);
  provide('tabs', tabs);
}

export function useTabsClient() {
  const tabs = inject('tabs');
  return { tabs };
}
