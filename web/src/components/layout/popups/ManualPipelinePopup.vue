<template>
  <Popup :open="open" @close="$emit('close')">
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
                  <i-la-times />
                </Button>
              </div>
            </div>
          </div>
        </InputField>
        <Button type="submit" :text="$t('repo.manual_pipeline.trigger')" />
      </form>
    </Panel>
  </Popup>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import Popup from '~/components/layout/Popup.vue';
import useApiClient from '~/compositions/useApiClient';
import { inject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();

const repo = inject('repo');

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
  const data = await usePaginate((page) => apiClient.getRepoBranches(repo.value.id, page));
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
