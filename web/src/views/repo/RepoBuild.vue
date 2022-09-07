<template>
  <Panel>
    <InputField :label="$t('repo.manual.select_branch')">
      <SelectField
        v-model="payload.branch"
        :options="branches"
        :disabled="loading"
        required
        class="dark:bg-dark-gray-700 bg-transparent text-color border-gray-200 dark:border-dark-400"
      />
    </InputField>
    <div class="flex flex-row mb-4 pb-4 items-center">
      <div class="">
        <h1 class="text-xl text-color">{{ $t('repo.manual.variables.title') }}</h1>
        <p class="text-sm text-color-alt">
          {{ $t('repo.manual.variables.desc') }}
        </p>
        <TextField
          v-model="tmpVar.key"
          class="m-2"
          :placeholder="$t('repo.manual.variables.key')"
          required
          :disabled="loading"
        />
        <TextField
          v-model="tmpVar.value"
          class="m-2"
          :placeholder="$t('repo.manual.variables.value')"
          required
          :disabled="loading"
        />
      </div>
      <Button
        class="ml-auto"
        start-icon="plus"
        :is-loading="loading"
        type="submit"
        :text="$t('repo.manual.variables.add')"
        @click="addVar"
      />
    </div>
    <div class="text-color">
      <div v-for="(v, k) in payload.variables" :key="k">
        <pre><span class="inline-block"><Button
          type="submit"
          text="X"
          class="inline-block"
          @click="deleteVar(k)"
        /></span>&nbsp;<span class="font-bold">{{ k }}</span>&#9;{{ v }}</pre>
      </div>
    </div>
    <br />
    <Button :is-loading="loading" type="submit" :text="$t('repo.manual.trigger')" @click="runManual" />
  </Panel>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';

const apiClient = useApiClient();

export default defineComponent({
  name: 'RepoBuild',

  components: {
    Panel,
    SelectField,
    InputField,
    TextField,
  },

  setup() {
    const route = useRoute();
    const router = useRouter();
    const branches = ref<{ text: string; value: string }[]>([]);
    const payload = ref<{ branch: string; variables: Record<string, string> }>({
      branch: 'main',
      variables: {
        MANUAL_BUILD: 'true',
      },
    });
    const loading = ref<boolean>(true);
    const tmpVar = ref<{ key: string; value: string }>({ key: '', value: '' });

    async function loadBranches() {
      const data = await apiClient.getRepoBranches(`${route.params.repoOwner}`, `${route.params.repoName}`);
      branches.value = data.map((e) => ({
        text: e,
        value: e,
      }));
      loading.value = false;
    }

    function addVar() {
      payload.value.variables[tmpVar.value.key] = tmpVar.value.value;
      tmpVar.value.key = '';
      tmpVar.value.value = '';
    }

    function deleteVar(key: string) {
      delete payload.value.variables[key];
    }

    async function runManual() {
      loading.value = true;
      const build = await apiClient.manualBuild(`${route.params.repoOwner}`, `${route.params.repoName}`, payload.value);

      router.push({
        name: 'repo-build',
        params: {
          repoOwner: `${route.params.repoOwner}`,
          repoName: `${route.params.repoName}`,
          buildId: build.number,
        },
      });
      loading.value = false;
    }

    onMounted(() => {
      loadBranches();
    });

    return {
      loading,
      branches,
      payload,
      tmpVar,
      addVar,
      deleteVar,
      runManual,
    };
  },
});
</script>
