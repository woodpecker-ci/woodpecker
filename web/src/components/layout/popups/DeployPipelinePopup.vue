<template>
  <Popup :open="open" @close="$emit('close')">
    <Panel v-if="!loading">
      <form @submit.prevent="triggerDeployPipeline">
        <span class="text-xl text-color">{{ $t('repo.deploy_pipeline.title', { pipelineId: pipelineNumber }) }}</span>
        <InputField :label="$t('repo.deploy_pipeline.enter_target')">
          <TextField v-model="payload.environment" required />
        </InputField>
        <InputField :label="$t('repo.deploy_pipeline.variables.title')">
          <span class="text-sm text-color-alt mb-2">{{ $t('repo.deploy_pipeline.variables.desc') }}</span>
          <div class="flex flex-col gap-2">
            <div v-for="(value, name) in payload.variables" :key="name" class="flex gap-4">
              <TextField :model-value="name" disabled />
              <TextField :model-value="value" disabled />
              <div class="w-34 flex-shrink-0">
                <Button color="red" class="ml-auto" @click="deleteVar(name)">
                  <i-la-times />
                </Button>
              </div>
            </div>
            <form class="flex gap-4" @submit.prevent="addPipelineVariable">
              <TextField
                v-model="newPipelineVariable.name"
                :placeholder="$t('repo.deploy_pipeline.variables.name')"
                required
              />
              <TextField
                v-model="newPipelineVariable.value"
                :placeholder="$t('repo.deploy_pipeline.variables.value')"
                required
              />
              <Button
                class="w-34 flex-shrink-0"
                start-icon="plus"
                type="submit"
                :text="$t('repo.deploy_pipeline.variables.add')"
              />
            </form>
          </div>
        </InputField>
        <Button type="submit" :text="$t('repo.deploy_pipeline.trigger')" />
      </form>
    </Panel>
  </Popup>
</template>

<script lang="ts" setup>
import { onMounted, ref, toRef } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import Popup from '~/components/layout/Popup.vue';
import useApiClient from '~/compositions/useApiClient';
import { inject } from '~/compositions/useInjectProvide';

const props = defineProps<{
  open: boolean;
  pipelineNumber: number;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();

const repo = inject('repo');

const router = useRouter();

const payload = ref<{ id: string; environment: string; variables: Record<string, string> }>({
  id: '',
  environment: '',
  variables: {},
});
const newPipelineVariable = ref<{ name: string; value: string }>({ name: '', value: '' });

const loading = ref(true);
onMounted(async () => {
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

const pipelineNumber = toRef(props, 'pipelineNumber');
async function triggerDeployPipeline() {
  loading.value = true;
  const newPipeline = await apiClient.deployPipeline(
    repo.value.owner,
    repo.value.name,
    pipelineNumber.value,
    payload.value,
  );

  emit('close');

  await router.push({
    name: 'repo-pipeline',
    params: {
      pipelineId: newPipeline.number,
    },
  });

  loading.value = false;
}
</script>
