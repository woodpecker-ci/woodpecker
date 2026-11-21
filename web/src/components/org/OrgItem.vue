<template>
  <router-link
    :to="{ name: 'org', params: { orgId: org.id } }"
    class="border-wp-background-400 dark:border-wp-background-100 bg-wp-background-200 dark:bg-wp-background-200 hover:bg-wp-control-neutral-100 dark:hover:bg-wp-control-neutral-200 flex cursor-pointer items-center gap-4 overflow-hidden rounded-md border p-4"
  >
    <span class="text-wp-text-100 text-lg">{{ org.name }}</span>
    <span class="text-wp-text-alt-100 ml-auto">{{
      $t('orgs.repo_count', { count: org.repoCount }, org.repoCount)
    }}</span>
  </router-link>
</template>

<script lang="ts" setup>
import type { OrgOverview } from '~/compositions/useOrgsFromRepos';

defineProps<{
  org: OrgOverview;
}>();
</script>
