<template>
  <Panel>
    <InputField :label="$t('manual.select_branch')">
      <SelectField
        v-model="payload.branch"
        :options="branches"
        :disabled="loading"
        required
        class="dark:bg-dark-gray-700 bg-transparent text-color border-gray-200 dark:border-dark-400"
      />
    </InputField>
    <div>
      <InputField :label="$t('manual.variable_key')">
        <TextField v-model="tmpVar.key" :placeholder="$t('manual.var_key')" required :disabled="loading" />
      </InputField>
      <InputField :label="$t('manual.variable_value')">
        <TextField v-model="tmpVar.value" :placeholder="$t('manual.var_value')" required :disabled="loading" />
      </InputField>
      <Button :is-loading="loading" type="submit" :text="$t('manual.add_variable')" @click="addVar" />
    </div>
    <br />
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
    <Button :is-loading="loading" type="submit" :text="$t('manual.launch_build')" @click="runManual" />
  </Panel>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';

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
    const branches = ref<{ text: string; value: string }[]>([]);
    const payload = ref<{ branch: string; variables: Record<string, string> }>({
      branch: 'main',
      variables: {
        MANUAL_BUILD: 'true',
      },
    });

    return {
      branches,
      payload,
    };
  },

  data: () => ({
    loading: true,
    tmpVar: {
      key: '',
      value: '',
    },
  }),

  mounted() {
    this.loadBranches();
  },

  methods: {
    async loadBranches() {
      const data = await apiClient.getRepoBranches(`${this.$route.params.repoOwner}`, `${this.$route.params.repoName}`);
      this.branches = data.map((e) => ({
        text: e,
        value: e,
      }));
      this.loading = false;
    },

    addVar() {
      this.payload.variables[this.tmpVar.key] = this.tmpVar.value;
      this.tmpVar.key = '';
      this.tmpVar.value = '';
    },

    deleteVar(key: string) {
      delete this.payload.variables[key];
    },

    async runManual() {
      this.loading = true;
      const build = await apiClient.manualBuild(
        `${this.$route.params.repoOwner}`,
        `${this.$route.params.repoName}`,
        this.payload,
      );

      this.$router.push({
        name: 'repo-build',
        params: {
          repoOwner: `${this.$route.params.repoOwner}`,
          repoName: `${this.$route.params.repoName}`,
          buildId: build.number,
        },
      });
      this.loading = false;
    },
  },
});
</script>
