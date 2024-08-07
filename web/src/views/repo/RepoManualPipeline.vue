<template>
  <Panel v-if="!loading">
    <form @submit.prevent="triggerManualPipeline">
      <span class="text-xl text-wp-text-100">{{ $t('repo.manual_pipeline.title') }}</span>
      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.select_branch')">
        <SelectField :id="id" v-model="payload.branch" :options="branches" required />
      </InputField>
      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.variables.title')">
        <span class="text-sm text-wp-text-alt-100 mb-2">{{ $t('repo.manual_pipeline.variables.desc') }}</span>
        <div class="flex flex-col gap-2">
          <div v-for="(_, i) in payload.variables" :key="i" class="flex gap-4">
            <TextField
              :id="id"
              v-model="payload.variables[i].name"
              :placeholder="$t('repo.manual_pipeline.variables.name')"
            />
            <TextField
              :id="id"
              v-model="payload.variables[i].value"
              :placeholder="$t('repo.manual_pipeline.variables.value')"
            />
            <div class="w-10 flex-shrink-0">
              <Button
                v-if="i !== payload.variables.length - 1"
                color="red"
                class="ml-auto"
                :title="$t('repo.manual_pipeline.variables.delete')"
                @click="deleteVar(i)"
              >
                <Icon name="remove" />
              </Button>
            </div>
          </div>
        </div>
      </InputField>
      <Button type="submit" :text="$t('repo.manual_pipeline.trigger')" />
    </form>
  </Panel>
  <div v-else class="flex justify-center text-wp-text-100">
    <Icon name="spinner" />
  </div>
</template>

<script lang="ts" setup>
import { useNotification } from '@kyvg/vue3-notification';
import type { Ref } from 'vue';
import { computed, onMounted, ref, inject as vueInject, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { inject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';
import type { RepoPermissions } from '~/lib/api/types';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();
const notifications = useNotification();
const i18n = useI18n();

const repo = inject('repo');
const repoPermissions = vueInject<Ref<RepoPermissions>>('repo-permissions');
if (!repoPermissions) {
  throw new Error('Unexpected: "repo" and "repoPermissions" should be provided at this place');
}

const router = useRouter();
const branches = ref<{ text: string; value: string }[]>([]);
const payload = ref<{ branch: string; variables: { name: string; value: string }[] }>({
  branch: 'main',
  variables: [
    {
      name: '',
      value: '',
    },
  ],
});

const pipelineOptions = computed(() => {
  const variables = Object.fromEntries(
    payload.value.variables.filter((e) => e.name !== '').map((item) => [item.name, item.value]),
  );
  return {
    ...payload.value,
    variables,
  };
});

const loading = ref(true);
onMounted(async () => {
  if (!repoPermissions.value.push) {
    notifications.notify({ type: 'error', title: i18n.t('repo.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }

  const data = await usePaginate((page) => apiClient.getRepoBranches(repo.value.id, { page }));
  branches.value = data.map((e) => ({
    text: e,
    value: e,
  }));
  loading.value = false;
});

watch(
  payload,
  () => {
    if (payload.value.variables[payload.value.variables.length - 1].name !== '') {
      payload.value.variables.push({
        name: '',
        value: '',
      });
    }
  },
  { deep: true },
);

function deleteVar(index: number) {
  payload.value.variables.splice(index, 1);
}

async function triggerManualPipeline() {
  loading.value = true;
  const pipeline = await apiClient.createPipeline(repo.value.id, pipelineOptions.value);

  emit('close');

  await router.push({
    name: 'repo-pipeline',
    params: {
      pipelineId: pipeline.number,
    },
  });

  loading.value = false;
}
</script>
