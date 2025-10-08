<template>
  <Popup :open="open" @close="$emit('close')">
    <Panel v-if="!loading">
      <form @submit.prevent="triggerDeployPipeline">
        <span class="text-wp-text-100 text-xl">{{
          $t('repo.deploy_pipeline.title', { pipelineId: pipelineNumber })
        }}</span>
        <InputField v-slot="{ id }" :label="$t('repo.deploy_pipeline.enter_target')">
          <TextField 
            :id="id" 
            v-model="payload.environment" 
            placeholder="Type or select target environment..."
            :list="`${id}-options`"
            required
          />
          <datalist :id="`${id}-options`">
            <option v-for="option in deployTargetOptions" :key="option.text" :value="option.value">
              {{ option.text }}
            </option>
          </datalist>
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
import { decode } from 'js-base64';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import Popup from '~/components/layout/Popup.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';

const props = defineProps<{
  open: boolean;
  pipelineNumber: string;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();
const repo = requiredInject('repo');
const router = useRouter();

const pipelineConfigs = requiredInject('pipeline-configs');

const decodedConfigs = computed(() => 
  pipelineConfigs.value?.map((config) => ({
    ...config,
    data: decode(config.data), // Decode base64 to readable YAML
  })) ?? []
);

// Extract CI_PIPELINE_DEPLOY_TARGET values
const deployTargetsFromConfigs = computed(() => {
  const targets = new Set<string>();
  
  decodedConfigs.value.forEach(config => {
    const yamlContent = config.data;
    // Look for evaluate: CI_PIPELINE_DEPLOY_TARGET == "somevalue" patterns
    const targetMatches = yamlContent.match(/CI_PIPELINE_DEPLOY_TARGET\s*==\s*["']([^"']+)["']/g);
    
    if (targetMatches) {
      targetMatches.forEach(match => {
        // Extract the value between quotes
        const valueMatch = match.match(/["']([^"']+)["']/);
        if (valueMatch && valueMatch[1]) {
          targets.add(valueMatch[1].trim());
        }
      });
    }
  });
  
  return Array.from(targets);
});

const deployTargetOptions = computed(() => {
  return deployTargetsFromConfigs.value.map(target => ({
    value: target,
    text: target
  }));
});

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
