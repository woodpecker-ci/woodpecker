import { computed, inject, onMounted, provide, Ref, ref } from 'vue';
import { useRoute } from 'vue-router';

export type Tab = {
  id: string;
  title: string;
};

export function useTabsProvider({
  activeTabProp,
  disableHashMode,
  updateActiveTabProp,
}: {
  activeTabProp: Ref<string | undefined>;
  updateActiveTabProp: (tab: string) => void;
  disableHashMode: Ref<boolean>;
}) {
  const route = useRoute();

  const tabs = ref<Tab[]>([]);
  const activeTab = ref<string>('');

  provide('tabs', tabs);
  provide(
    'disable-hash-mode',
    computed(() => disableHashMode.value),
  );
  provide(
    'active-tab',
    computed({
      get: () => activeTab.value,
      set: (value) => {
        activeTab.value = value;
        updateActiveTabProp(value);
      },
    }),
  );

  onMounted(() => {
    if (activeTabProp.value) {
      activeTab.value = activeTabProp.value;
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
  const disableHashMode = inject<Ref<boolean>>('disable-hash-mode');
  const activeTab = inject<Ref<string>>('active-tab');

  if (activeTab === undefined || tabs === undefined || disableHashMode === undefined) {
    throw new Error('Please use this "useTabsClient" composition inside a component running "useTabsProvider".');
  }

  return { activeTab, tabs, disableHashMode };
}
