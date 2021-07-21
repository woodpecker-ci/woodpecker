<template>
  <div v-if="build && repo">
    <FluidContainer class="flex border-b mb-4 items-start items-center">
      <Breadcrumbs
        :paths="[
          repo.owner,
          { name: repo.name, link: { name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } } },
          {
            name: `Build #${buildId}`,
            link: { name: 'repo-build', params: { repoOwner: repo.owner, repoName: repo.name, buildId } },
          },
        ]"
      />
      <span class="text-xl mx-auto">{{ message }}</span>
      <BuildStatusIcon :build="build" class="ml-auto" />
      <Button class="ml-4" text="Cancel" />
    </FluidContainer>

    <FluidContainer>
      <div class="flex justify-evenly text-gray-500">
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

      <div class="flex mt-4 w-full bg-gray-600 min-h-0 rounded-md overflow-hidden">
        <div class="flex flex-col w-3/12 text-white">
          <div v-for="proc in build.procs" :key="proc.id">
            <div class="p-2">{{ proc.name }}</div>
            <div
              v-for="job in proc.children"
              :key="job.pid"
              class="flex p-2 pl-6 cursor-pointer items-center"
              :class="{ 'bg-gray-800': selectedJob === job.pid }"
              @click="selectedJob = job.pid"
            >
              <div v-if="['success'].includes(job.state)" class="w-2 h-2 bg-status-success rounded-full" />
              <div v-if="['pending', 'skipped'].includes(job.state)" class="w-2 h-2 bg-status-pending rounded-full" />
              <div
                v-if="['killed', 'error', 'failure', 'blocked', 'declined'].includes(job.state)"
                class="w-2 h-2 bg-status-error rounded-full"
              />
              <div v-if="['started', 'running'].includes(job.state)" class="w-2 h-2 bg-status-running rounded-full" />
              <span class="ml-2">{{ job.name }}</span>
              <span class="ml-auto" v-if="job.start_time !== undefined">{{ jobDuration(job) }}</span>
            </div>
          </div>
        </div>

        <BuildLogs :build="build" class="w-9/12 flex-grow" />
      </div>
    </FluidContainer>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, Ref, ref, toRef, watch } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo, Build, BuildJob } from '~/lib/api/types';
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
import BuildLogs from '~/components/repo/BuildLogs.vue';
import useBuild from '~/compositions/useBuild';
import { durationAsNumber, prettyDuration } from '~/utils/duration';

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
    BuildLogs,
  },

  props: {
    buildId: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const apiClient = useApiClient();

    const buildId = toRef(props, 'buildId');
    const repo = computed(() => inject<Ref<Repo>>('repo')?.value || null);
    const build = ref<Build | undefined>();
    const { since, duration, message } = useBuild(build);
    const selectedJob = ref(5);

    async function loadBuild(): Promise<void> {
      if (!repo.value) {
        return;
      }

      build.value = await apiClient.getBuild(repo.value.owner, repo.value.name, buildId.value);
    }

    function jobDuration(job: BuildJob): string {
      const start = job.start_time || 0;
      const end = job.end_time || 0;

      if (end === 0 && start === 0) {
        return '-';
      }

      if (end === 0) {
        return durationAsNumber(Date.now() - start * 1000);
      }

      return durationAsNumber((end - start) * 1000);
    }

    onMounted(loadBuild);
    watch([repo, buildId], loadBuild);

    return { build, since, duration, repo, message, selectedJob, jobDuration };
  },
});
</script>
