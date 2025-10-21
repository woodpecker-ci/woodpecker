import { defineStore } from 'pinia';
import { computed, reactive } from 'vue';
import type { Ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import type { Forge } from '~/lib/api/types';

export const useForgeStore = defineStore('forges', () => {
  const apiClient = useApiClient();

  const forges = reactive<Map<number, Forge>>(new Map());

  async function loadForge(forgeId: number): Promise<Forge> {
    const forge = await apiClient.getForge(forgeId);
    forges.set(forge.id, forge);
    return forge;
  }

  async function getForge(forgeId: number): Promise<Ref<Forge | undefined>> {
    if (!forges.has(forgeId)) {
      await loadForge(forgeId);
    }

    return computed(() => forges.get(forgeId));
  }

  return {
    getForge,
    loadForge,
  };
});
