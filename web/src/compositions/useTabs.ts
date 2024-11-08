import { computed, inject, onMounted, provide, ref, type Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import type { IconNames } from '~/components/atomic/Icon.vue';

export interface Tab {
  to: string;
  alternativeRoute?: string;
  title: string;
  icon?: IconNames;
  iconClass?: string;
}

export function useTabsProvider() {
  const route = useRoute();
  const router = useRouter();

  const alternativeRoute: Record<string, string> = {};

  const tabs = ref<Tab[]>([]);

  const activeTab = computed({
    get() {
      if (route.name !== undefined) {
        if (route.name.toString() in alternativeRoute) {
          return alternativeRoute[route.name.toString()];
        }
        return route.name.toString()
      }
      return tabs.value[0].to
    },
    set(tab) {
      router.push({ name: tab }).catch(console.error);
    },
  });

  provide('tabs', tabs);
  provide('active-tab', activeTab);

  onMounted(() => {
    for (const i of tabs.value) {
      if (i.alternativeRoute !== undefined) {
        alternativeRoute[i.alternativeRoute] = i.to;
      }
    }
  });
}

export function useTabsClient() {
  const tabs = inject<Ref<Tab[]>>('tabs');
  const activeTab = inject<Ref<string>>('active-tab');

  if (activeTab === undefined || tabs === undefined) {
    throw new Error('Please use this "useTabsClient" composition inside a component running "useTabsProvider".');
  }

  return { activeTab, tabs };
}
