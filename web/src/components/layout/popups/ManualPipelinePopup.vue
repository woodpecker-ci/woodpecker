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
            <div v-for="(value, name) in payload.variables" :key="name" class="flex gap-4">
              <TextField :id="id" :model-value="name" disabled />
              <TextField :id="id" :model-value="value" disabled />
              <div class="w-34 flex-shrink-0">
                <Button color="red" class="ml-auto" @click="deleteVar(name)">
                  <i-la-times />
                </Button>
              </div>
            </div>
            <form class="flex gap-4" @submit.prevent="addPipelineVariable">
              <TextField
                :id="id"
                v-model="newPipelineVariable.name"
                :placeholder="$t('repo.manual_pipeline.variables.name')"
                required
              />
              <TextField
                :id="id"
                v-model="newPipelineVariable.value"
                :placeholder="$t('repo.manual_pipeline.variables.value')"
                required
              />
              <Button
                class="w-34 flex-shrink-0"
                start-icon="plus"
                type="submit"
                :text="$t('repo.manual_pipeline.variables.add')"
              />
            </form>
          </div>
        </InputField>
        <Button type="submit" :text="$t('repo.manual_pipeline.trigger')" />
      </form>
    </Panel>
  </Popup>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
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
const payload = ref<{ branch: string; variables: Record<string, string> }>({
  branch: 'main',
  variables: {},
});
const newPipelineVariable = ref<{ name: string; value: string }>({ name: '', value: '' });

const loading = ref(true);
onMounted(async () => {
  const data = await usePaginate((page) => apiClient.getRepoBranches(repo.value.id, page));
  branches.value = data.map((e) => ({
    text: e,
    value: e,
  }));
  loading.value = false;
});

function addPipelineVariable() {
  if (!newPipelineVariable.value.name || !newPipelineVariable.value.value) {
    return;
  }
  payload.value.variables[newPipelineVariable.value.name] = newPipelineVariable.value.value;
  newPipelineVariable.value.name = '';
  newPipelineVariable.value.value = '';
}

function deleteVar(key: string) {
  delete payload.value.variables[key];
}

async function triggerManualPipeline() {
  loading.value = true;
  const pipeline = await apiClient.createPipeline(repo.value.id, payload.value);

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
