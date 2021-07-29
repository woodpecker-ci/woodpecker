<template>
  <template v-if="build && repo">
    <FluidContainer class="flex border-b mb-4 items-start items-center">
      <Breadcrumbs
        :paths="[
          { name: 'Repositories', link: { name: 'home' } },
          {
            name: `${repo.owner} / ${repo.name}`,
            link: { name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } },
          },
          {
            name: `Build #${buildId}`,
            link: { name: 'repo-build', params: { repoOwner: repo.owner, repoName: repo.name, buildId } },
          },
        ]"
      />
      <BuildStatusIcon :build="build" class="flex ml-auto" />
      <Button class="ml-4" text="Cancel" />
    </FluidContainer>

    <FluidContainer class="p-0 flex flex-col flex-grow">
      <span class="text-xl mx-auto mb-4">{{ message }}</span>

      <div class="flex mx-auto space-x-16 text-gray-500">
        <div class="flex space-x-2 items-center">
          <icon-commit />
          <a class="text-link" :href="build.link_url" target="_blank">{{ build.commit.slice(0, 10) }}</a>
        </div>
        <div class="flex space-x-2 items-center">
          <icon-branch />
          <span>{{ build.branch }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <icon-since />
          <span>{{ since }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <icon-duration />
          <span>{{ duration }}</span>
        </div>
      </div>

      <BuildProcs :build="build" v-model:selected-proc-id="selectedProcId" />
    </FluidContainer>
  </template>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, Ref, toRef, watch } from 'vue';
import BuildStore from '~/store/builds';
import { Repo } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import BuildItem from '~/components/repo/BuildItem.vue';
import Breadcrumbs from '~/components/layout/Breadcrumbs.vue';
import IconDuration from 'virtual:vite-icons/ic/sharp-timelapse';
import IconSince from 'virtual:vite-icons/mdi/clock-time-eight-outline';
import IconBranch from 'virtual:vite-icons/mdi/source-branch';
import IconGithub from 'virtual:vite-icons/mdi/github';
import IconCommit from 'virtual:vite-icons/mdi/source-commit';
import BuildStatusIcon from '~/components/repo/BuildStatusIcon.vue';
import useBuild from '~/compositions/useBuild';
import { useRouter, useRoute } from 'vue-router';
import BuildProcs from '~/components/repo/BuildProcs.vue';

export default defineComponent({
  name: 'Build',

  components: {
    FluidContainer,
    Button,
    BuildItem,
    Breadcrumbs,
    IconDuration,
    IconSince,
    IconBranch,
    IconGithub,
    IconCommit,
    BuildStatusIcon,
    BuildProcs,
  },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
    repoName: {
      type: String,
      required: true,
    },
    buildId: {
      type: String,
      required: true,
    },
    procId: {
      type: String,
      required: false,
    },
  },

  setup(props) {
    const router = useRouter();
    const route = useRoute();

    const buildStore = BuildStore();
    const buildId = toRef(props, 'buildId');
    const repoOwner = toRef(props, 'repoOwner');
    const repoName = toRef(props, 'repoName');
    const repo = inject<Ref<Repo>>('repo');
    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    const build = buildStore.getBuild(repoOwner, repoName, buildId);
    const { since, duration, message } = useBuild(build);
    const procId = toRef(props, 'procId');
    const selectedProcId = computed({
      get() {
        if (procId.value) {
          return parseInt(procId.value);
        }

        if (!build.value || !build.value.procs || !build.value.procs[0].children) {
          return null;
        }

        return build.value.procs[0].children[0].pid;
      },
      set(selectedProcId: number | null) {
        if (!selectedProcId) {
          return;
        }

        router.replace({ params: { ...route.params, procId: selectedProcId } });
      },
    });

    async function loadBuild(): Promise<void> {
      if (!repo) {
        return;
      }

      await buildStore.loadBuild(repo.value.owner, repo.value.name, parseInt(buildId.value));
    }

    onMounted(loadBuild);
    watch([repo, buildId], loadBuild);

    return { selectedProcId, build, since, duration, repo, message };
  },
});
</script>
