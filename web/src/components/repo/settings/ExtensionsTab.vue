<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-gray-500">Extensions</h1>
    </div>

    <div class="flex flex-col">
      {{ repo }}
    </div>
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, Ref } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'ExtensionsTab',

  components: { Button, Panel },

  setup() {
    const apiClient = useApiClient();
    const router = useRouter();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');

    return {
      repo,
    };
  },
});
</script>
