<template>
  <Popup :open="open" @close="$emit('close')">
    <Panel v-if="!loading">
      <form @submit.prevent="triggerDeployPipeline">
        <span class="text-xl text-wp-text-100">{{
          $t('repo.deploy_pipeline.title', { pipelineId: pipelineNumber })
        }}</span>
        <InputField v-slot="{ id }" :label="$t('repo.deploy_pipeline.enter_target')">
          <TextField :id="id" v-model="payload.environment" required />
        </InputField>
        <InputField v-slot="{ id }" :label="$t('repo.deploy_pipeline.variables.title')">
          <span class="text-sm text-wp-text-alt-100 mb-2">{{ $t('repo.deploy_pipeline.variables.desc') }}</span>
          <div class="flex flex-col gap-2">
            <div v-for="(_, i) in payload.variables" :key="i" class="flex gap-4">
              <TextField
                :id="id"
                v-model="payload.variables[i].name"
                :placeholder="$t('repo.deploy_pipeline.variables.name')"
              />
              <TextField
                :id="id"
                v-model="payload.variables[i].value"
                :placeholder="$t('repo.deploy_pipeline.variables.value')"
              />
              <div class="w-10 flex-shrink-0">
                <Button
                  v-if="i !== payload.variables.length - 1"
                  color="red"
                  class="ml-auto"
                  :title="$t('repo.deploy_pipeline.variables.delete')"
                  @click="deleteVar(i)"
                >
                  <i-la-times />
                </Button>
              </div>
            </div>
          </div>
        </InputField>
        <Button type="submit" :text="$t('repo.deploy_pipeline.trigger')" />
      </form>
    </Panel>
  </Popup>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, toRef, watch } from 'vue';
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
  pipelineNumber: string;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();

const repo = inject('repo');

const router = useRouter();

const payload = ref<{ id: string; environment: string; variables: { name: string; value: string }[] }>({
  id: '',
  environment: '',
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
