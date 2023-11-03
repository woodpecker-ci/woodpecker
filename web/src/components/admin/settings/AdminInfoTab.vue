<template>
  <Settings :title="$t('admin.settings.info.info')" :desc="$t('admin.settings.info.desc')">
    <div class="flex flex-col items-center gap-4">
      <WoodpeckerLogo class="w-48 h-48" />
      <p class="text-xl">
        You are running woodpecker version <span class="font-bold">{{ version?.current }}</span>
      </p>
      <span v-if="version?.needsUpdate" class="text-red-400"
        >Please update your Woodpecker instance to {{ version.latest }}</span
      >
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import WoodpeckerLogo from '~/assets/logo.svg?component';
import Settings from '~/components/layout/Settings.vue';
import { useVersion } from '~/compositions/useVersion';

const version = useVersion();
</script>
