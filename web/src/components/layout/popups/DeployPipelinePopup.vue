<template>
  <Popup :open="open" @close="$emit('close')">
    <Panel v-if="!loading">
      <form @submit.prevent="triggerDeployPipeline">
        <span class="text-wp-text-100 text-xl">{{
          $t('repo.deploy_pipeline.title', { pipelineId: pipelineNumber })
        }}</span>
        <InputField v-slot="{ id }" :label="$t('repo.deploy_pipeline.enter_target')">
          <TextField :id="id" v-model="payload.environment" required />
        </InputField>
        <InputField v-slot="{ id }" :label="$t('repo.deploy_pipeline.enter_task')">
          <TextField :id="id" v-model="payload.task" />
        </InputField>
        <InputField v-slot="{ id }" :label="$t('repo.deploy_pipeline.variables.title')">
          <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.deploy_pipeline.variables.desc') }}</span>
          <KeyValueEditor
            :id="id"
            v-model="payload.variables"
            :key-placeholder="$t('repo.deploy_pipeline.variables.name')"
            :value-placeholder="$t('repo.deploy_pipeline.variables.value')"
            :delete-title="$t('repo.deploy_pipeline.variables.delete')"
            @update:is-valid="isVariablesValid = $event"
          />
        </InputField>
        <Button type="submit" :text="$t('repo.deploy_pipeline.trigger')" :disabled="!isFormValid" />
      </form>
    </Panel>
  </Popup>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, toRef } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import Popup from '~/components/layout/Popup.vue';
import useApiClient from '~/compositions/useApiClient';
import { inject } from '~/compositions/useInjectProvide';

const props = defineProps<{
  open: boolean;
  pipelineNumber: string;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();
const repo = inject('repo');
const router = useRouter();

const payload = ref<{
  id: string;
  environment: string;
  task: string;
  variables: Record<string, string>;
}>({
  id: '',
  environment: '',
  task: '',
  variables: {},
});

const isVariablesValid = ref(true);

const isFormValid = computed(() => {
  return payload.value.environment !== '' && isVariablesValid.value;
});

const pipelineOptions = computed(() => ({
  ...payload.value,
  variables: payload.value.variables,
}));

const loading = ref(true);
onMounted(async () => {
  loading.value = false;
});

const pipelineNumber = toRef(props, 'pipelineNumber');
async function triggerDeployPipeline() {
  loading.value = true;
  const newPipeline = await apiClient.deployPipeline(repo.value.id, pipelineNumber.value, pipelineOptions.value);

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
