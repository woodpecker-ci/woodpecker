import { inject, provide, ref, type Ref } from 'vue';
import type { RouteLocationRaw } from 'vue-router';

import type { IconNames } from '~/components/atomic/Icon.vue';

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
  const tabs = inject<Ref<Tab[]>>('tabs');

  if (tabs === undefined) {
    throw new Error('Please use this "useTabsClient" composition inside a component running "useTabsProvider".');
  }

  return { tabs };
}
