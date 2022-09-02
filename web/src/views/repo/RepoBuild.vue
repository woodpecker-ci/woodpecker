<template>
  <Panel>
    <InputField :label="$t('manual.select_branch')">
      <SelectField
        v-model="this.payload.branch"
        :options="this.branches"
        :disabled="loading"
        required
        class="bg-transparent text-color"
      />
    </InputField>
    <div>
      <InputField :label="$t('manual.variable_key')">
        <TextField
          v-model="this.tmpVar.key"
          :placeholder="$t('manual.var_key')"
          required
          :disabled="loading"
        />
      </InputField>
      <InputField :label="$t('manual.variable_value')">
        <TextField
          v-model="this.tmpVar.value"
          :placeholder="$t('manual.var_value')"
          required
          :disabled="loading"
        />
      </InputField>
      <Button
        :is-loading="loading"
        type="submit"
        :text="$t('manual.add_variable')"
        @click="addVar()"
      />
    </div>
    <br>
    <div class="text-color">
      <div v-for="(v, k) in this.payload.variables" :key="k">
      <pre><Button
        type="submit"
        text="X"
        @click="deleteVar(k)"
        style="display: inline-block;"
      />&nbsp;<span class="font-bold">{{ k }}</span>&#9;{{ v }}</pre>
      </div>
    </div>
    <br>
    <Button
      :is-loading="loading"
      type="submit"
      :text="$t('manual.launch_build')"
      @click="runManual()"
    />
  </Panel>
</template>

<script lang="ts">
import {defineComponent, inject, onMounted, Ref} from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import useApiClient from '~/compositions/useApiClient';

import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import SelectField from "../../components/form/SelectField.vue";
import Panel from "../../components/layout/Panel.vue";

const apiClient = useApiClient();

export default defineComponent({
  name: 'RepoBuild',

  components: {
    Panel,
    SelectField,
    IconButton,
    InputField,
    TextField
  },

  setup() {

  },

  mounted() {
    apiClient.getRepoBranches(`${this.$route.params.repoOwner}`, `${this.$route.params.repoName}`).then((b) => {
      this.branches = b.map((e) => {
        return {
          text: e,
          value: e
        }
      })
      this.loading = false
    })
  },

  data: () => {
    return {
      loading: true,
      branches: [],
      payload: {
        branch: 'main',
        variables: {
          MANUAL_BUILD: "true"
        }
      },
      tmpVar: {
        key: "",
        value: ""
      }
    }
  },

  methods: {
    addVar() {
      this.payload.variables[this.tmpVar.key] = this.tmpVar.value
      this.tmpVar.key = ''
      this.tmpVar.value = ''
    },
    deleteVar(key: string) {
      console.log(key)
      delete this.payload.variables[key]
    },
    runManual() {
      this.loading = true
      apiClient
        .manualBuild(`${this.$route.params.repoOwner}`, `${this.$route.params.repoName}`, this.payload)
        .then((build) => {
          this.$router.push({
            name: 'repo-build',
            params: {
              repoOwner: `${this.$route.params.repoOwner}`,
              repoName: `${this.$route.params.repoName}`,
              buildId: build.number
            }
          });
        }).catch((error) => {
        alert(JSON.stringify(error))
      }).finally(() => {
        this.loading = false
      });
    }
  }
});
</script>
