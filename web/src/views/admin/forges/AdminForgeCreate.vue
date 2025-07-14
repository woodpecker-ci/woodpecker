<template>
  <Settings :title="$t('forges')" :description="$t('forges_desc')">
    <template #headerActions>
      <Button
        :text="$t('show_forges')"
        start-icon="back"
        :to="{ name: 'admin-settings-forges' }"
      />
    </template>

    <AdminForgeForm v-model:forge="forge" :is-saving="isSaving" is-new @submit="saveForge" />
  </Settings>
</template>

<script lang="ts" setup>
import {  ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import AdminForgeForm from '~/components/admin/settings/forges/AdminForgeForm.vue';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import type { Forge, } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();
const router = useRouter();

const forge = ref<Partial<Forge>>({});

const { doSubmit: saveForge, isLoading: isSaving } = useAsyncAction(async () => {
  if (!forge.value) {
    throw new Error("Unexpected: Can't get forge");
  }

  forge.value = await apiClient.createForge(forge.value);
  notifications.notify({
    title: t('forge_created'),
    type: 'success',
  });

  await router.push({ name: 'admin-settings-forge', params: { forgeId: forge.value.id } });
});
</script>
