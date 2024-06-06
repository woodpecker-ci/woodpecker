<template>
  <div />
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import useApiClient from '~/compositions/useApiClient';

const props = defineProps<{
  orgName: string;
}>();
const apiClient = useApiClient();
const route = useRoute();
const router = useRouter();

onMounted(async () => {
  const org = await apiClient.lookupOrg(props.orgName);

  const path = route.path.replace(`/org/${props.orgName}`, `/orgs/${org?.id}`);

  await router.replace({ path });
});
</script>
