<template>
  <Panel>
    <template v-if="!loading">
      <span class="text-xl text-color">{{ $t('repo.manual.title') }}</span>
      <InputField :label="$t('repo.manual.select_branch')">
        <SelectField
          v-model="payload.branch"
          :options="branches"
          required
          class="dark:bg-dark-gray-700 bg-transparent text-color border-gray-200 dark:border-dark-400"
        />
      </InputField>
      <InputField :label="$t('repo.manual.variables.title')">
        <span class="text-sm text-color-alt mb-2">{{ $t('repo.manual.variables.desc') }}</span>
        <div class="flex flex-col gap-2">
          <div v-for="(value, name) in payload.variables" :key="name" class="flex gap-4">
            <TextField :model-value="name" disabled />
            <TextField :model-value="value" disabled />
            <div class="w-34 flex-shrink-0">
              <Button type="submit" class="ml-auto" @click="deleteVar(name)">
                <i-la-times />
              </Button>
            </div>
          </div>
          <form class="flex gap-4" @submit.prevent="addPipelineVariable">
            <TextField v-model="newPipelineVariable.name" :placeholder="$t('repo.manual.variables.name')" required />
            <TextField v-model="newPipelineVariable.value" :placeholder="$t('repo.manual.variables.value')" required />
            <Button
              class="w-34 flex-shrink-0"
              start-icon="plus"
              type="submit"
              :text="$t('repo.manual.variables.add')"
            />
          </form>
        </div>
      </InputField>
      <Button type="submit" :text="$t('repo.manual.trigger')" @click="triggerManualPipeline" />
    </template>
  </Panel>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { inject } from '~/compositions/useInjectProvide';

const apiClient = useApiClient();

const repo = inject('repo');

const router = useRouter();
const branches = ref<{ text: string; value: string }[]>([]);
const payload = ref<{ branch: string; variables: Record<string, string> }>({
  branch: 'main',
  variables: {
    MANUAL_BUILD: 'true',
  },
});
const newPipelineVariable = ref<{ name: string; value: string }>({ name: '', value: '' });

const loading = ref(true);
onMounted(async () => {
  const data = await apiClient.getRepoBranches(repo.value.owner, repo.value.name);
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
  const build = await apiClient.createBuild(repo.value.owner, repo.value.name, payload.value);

  await router.push({
    name: 'repo-build',
    params: {
      buildId: build.number,
    },
  });

  loading.value = false;
}
</script>
