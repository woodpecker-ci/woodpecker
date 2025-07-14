<template>
  <Settings :title="$t('forges')" :description="$t('forges_desc')">
    <template #headerActions>
      <Button
        :text="$t('show_forges')"
        start-icon="back"
        :to="{ name: 'admin-settings-forges' }"
      />
    </template>

    <AdminForgeForm v-if="forge" v-model:forge="forge" :is-saving="isSaving" @submit="saveForge" />
</Settings>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import AdminForgeForm from '~/components/admin/settings/forges/AdminForgeForm.vue';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import type { Forge } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();
const route = useRoute();

const forgeId = computed(() => Number.parseInt(route.params.forgeId.toString(), 10));
const forge = ref<Forge>();

async function load() {
  forge.value = await apiClient.getForge(forgeId.value);
}

onMounted(load);
watch(forgeId, load);

const { doSubmit: saveForge, isLoading: isSaving } = useAsyncAction(async () => {
  if (!forge.value) {
    throw new Error("Unexpected: Can't get forge");
  }

  await apiClient.updateForge(forge.value);
  notifications.notify({
    title: t('forge_saved'),
    type: 'success',
  });

  await load(); // reload
});
</script>
