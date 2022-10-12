<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-color">{{ $t('repo.settings.badge.badge') }}</h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>
    </div>

    <InputField :label="$t('repo.settings.badge.type')">
      <SelectField
        v-model="badgeType"
        :options="[
          {
            value: 'url',
            text: $t('repo.settings.badge.type_url'),
          },
          {
            value: 'markdown',
            text: $t('repo.settings.badge.type_markdown'),
          },
          {
            value: 'html',
            text: $t('repo.settings.badge.type_html'),
          },
        ]"
        required
        class="dark:bg-dark-gray-700 bg-transparent text-color border-gray-200 dark:border-dark-400"
      />
    </InputField>
    <InputField :label="$t('repo.settings.badge.branch')">
      <TextField
        v-model="branch"
        :placeholder="defaultBranch"
        class="dark:bg-dark-gray-700 bg-transparent text-color border-gray-200 dark:border-dark-400"
      />
    </InputField>

    <div v-if="badgeContent" class="flex flex-col space-y-4">
      <div>
        <pre class="box">{{ badgeContent }}</pre>
      </div>
    </div>
  </Panel>
</template>

<script lang="ts">
import { useStorage } from '@vueuse/core';
import { computed, defineComponent, inject, Ref, ref } from 'vue';

import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BadgeTab',

  components: { Panel, InputField, SelectField, TextField },

  setup() {
    const repo = inject<Ref<Repo>>('repo');

    const badgeType = useStorage('last-badge-type', 'markdown');

    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    const defaultBranch = computed(() => repo.value.default_branch);
    const branch = ref<string>('');

    const baseUrl = `${window.location.protocol}//${window.location.hostname}${
      window.location.port ? `:${window.location.port}` : ''
    }`;
    const badgeUrl = computed(
      () =>
        `/api/badges/${repo.value.owner}/${repo.value.name}/status.svg${
          branch.value !== '' ? `?branch=${branch.value}` : ''
        }`,
    );
    const repoUrl = computed(
      () => `/${repo.value.owner}/${repo.value.name}${branch.value !== '' ? `/branches/${branch.value}` : ''}`,
    );

    const badgeContent = computed(() => {
      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      if (badgeType.value === 'url') {
        return `${baseUrl}${badgeUrl.value}`;
      }

      if (badgeType.value === 'markdown') {
        return `[![status-badge](${baseUrl}${badgeUrl.value})](${baseUrl}${repoUrl.value})`;
      }

      if (badgeType.value === 'html') {
        return `<a href="${baseUrl}${repoUrl.value}" target="_blank">\n  <img src="${baseUrl}${badgeUrl.value}" alt="status-badge" />\n</a>`;
      }

      return '';
    });

    return { badgeType, defaultBranch, branch, badgeContent, badgeUrl };
  },
});
</script>

<style scoped>
.box {
  @apply bg-gray-500 p-2 rounded-md text-white break-words dark:bg-dark-400 dark:text-gray-400;
  white-space: pre-wrap;
}
</style>
