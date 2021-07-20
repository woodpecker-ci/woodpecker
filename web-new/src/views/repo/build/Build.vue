<template>
  <div v-if="build">
    <FluidContainer class="flex border-b mb-4 items-start items-center">
      <Breadcrumbs
        :paths="[
          repoOwner,
          { name: repoId, link: { name: 'repo', params: { repoOwner, repoId } } },
          { name: `Build #${buildId}`, link: { name: 'repo-build', params: { repoOwner, repoId, buildId } } },
        ]"
      />
      <Button class="ml-auto" text="Cancel" />
    </FluidContainer>
    <FluidContainer>
      <pre>
        {{ build }}
      </pre>
    </FluidContainer>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRef, watch } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo, Build } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import { useRouter } from 'vue-router';
import useNotifications from '~/compositions/useNotifications';
import BuildItem from '~/components/repo/BuildItem.vue';
import Breadcrumbs from '~/components/layout/Breadcrumbs.vue';

export default defineComponent({
  name: 'Repo',

  components: { FluidContainer, Button, BuildItem, Breadcrumbs },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
    repoId: {
      type: String,
      required: true,
    },
    buildId: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const apiClient = useApiClient();

    const repoOwner = toRef(props, 'repoOwner');
    const repoId = toRef(props, 'repoId');
    const buildId = computed(() => parseInt(toRef(props, 'buildId').value));

    const build = ref<Build | undefined>();

    onMounted(async () => {
      build.value = await apiClient.getBuild(repoOwner.value, repoId.value, buildId.value);
    });

    watch([repoOwner, repoId, buildId], async () => {
      build.value = await apiClient.getBuild(repoOwner.value, repoId.value, buildId.value);
    });

    return { build };
  },
});
</script>
