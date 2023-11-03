import { inject, onMounted, provide, Ref, ref } from 'vue';
import { useRoute } from 'vue-router';

export type Tab = {
  id: string;
  title: string;
};

export function useTabsProvider({
  activeTab,
  disableUrlHashMode,
}: {
  activeTab: Ref<string | undefined>;
  disableUrlHashMode: Ref<boolean>;
}) {
  const route = useRoute();

  const tabs = ref<Tab[]>([]);

  provide('tabs', tabs);
  provide('disable-url-hash-mode', disableUrlHashMode);
  provide('active-tab', activeTab);

  onMounted(() => {
    if (activeTab.value !== undefined) {
      return;
    }

    const hashTab = route.hash.replace(/^#/, '');
    if (hashTab) {
      activeTab.value = hashTab;
      return;
    }

    activeTab.value = tabs.value[0].id;
  });
}

export function useTabsClient() {
  const tabs = inject<Ref<Tab[]>>('tabs');
  const disableUrlHashMode = inject<Ref<boolean>>('disable-url-hash-mode');
  const activeTab = inject<Ref<string>>('active-tab');

  if (activeTab === undefined || tabs === undefined || disableUrlHashMode === undefined) {
    throw new Error('Please use this "useTabsClient" composition inside a component running "useTabsProvider".');
  }

  return { activeTab, tabs, disableUrlHashMode };
}
