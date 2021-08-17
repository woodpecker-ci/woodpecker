<template>
  <Panel>
    <span class="text-lg border-b-2 w-full">Registries</span>
  </Panel>
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';
import { useRouter } from 'vue-router';
import useNotifications from '~/compositions/useNotifications';
import Panel from '~/components/layout/Panel.vue';

export default defineComponent({
  name: 'RegistriesTab',

  components: { Button, Panel },

  setup() {
    const apiClient = useApiClient();
    const router = useRouter();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');

    return { repo };
  },
});
</script>
