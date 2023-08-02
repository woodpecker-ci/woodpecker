<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-wp-background-100">
      <h1 class="text-xl ml-2 text-wp-text-100">{{ $t('repo.settings.badge.badge') }}</h1>
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
      />
    </InputField>
    <InputField :label="$t('repo.settings.badge.branch')">
      <SelectField v-model="branch" :options="branches" required />
    </InputField>

    <div v-if="badgeContent" class="flex flex-col space-y-4">
      <div>
        <pre class="code-box">{{ badgeContent }}</pre>
      </div>
    </div>
  </Panel>
</template>

<script lang="ts">
import { useStorage } from '@vueuse/core';
import { computed, defineComponent, inject, onMounted, Ref, ref, watch } from 'vue';

import { SelectOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { usePaginate } from '~/compositions/usePaginate';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BadgeTab',

  components: { Panel, InputField, SelectField },

  setup() {
    const apiClient = useApiClient();
    const repo = inject<Ref<Repo>>('repo');

    const badgeType = useStorage('last-badge-type', 'markdown');

    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    const defaultBranch = computed(() => repo.value.default_branch);
    const branches = ref<SelectOption[]>([]);
    const branch = ref<string>('');

    async function loadBranches() {
      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      branches.value = (await usePaginate((page) => apiClient.getRepoBranches(repo.value.id, page)))
        .map((b) => ({
          value: b,
          text: b,
        }))
        .filter((b) => b.value !== defaultBranch.value);
      branches.value.unshift({
        value: '',
        text: defaultBranch.value,
      });
    }

    const baseUrl = `${window.location.protocol}//${window.location.hostname}${
      window.location.port ? `:${window.location.port}` : ''
    }`;
    const badgeUrl = computed(
      () => `/api/badges/${repo.value.id}/status.svg${branch.value !== '' ? `?branch=${branch.value}` : ''}`,
    );
    const repoUrl = computed(
      () => `/repos/${repo.value.id}${branch.value !== '' ? `/branches/${encodeURIComponent(branch.value)}` : ''}`,
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

    onMounted(() => {
      loadBranches();
    });

    watch(repo, () => {
      loadBranches();
    });

    return { badgeType, branches, branch, badgeContent, badgeUrl };
  },
});
</script>
