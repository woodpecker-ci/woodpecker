<template>
  <Settings :title="$t('info')">
    <div class="flex flex-col items-center gap-4">
      <WoodpeckerLogo class="w-48 h-48" />
      <p class="text-xl">
        {{ $t('running_version') }} <span class="font-bold">{{ version?.current }}</span>
      </p>
      <span v-if="version?.needsUpdate" class="text-int-wp-state-error-100">
        {{ $t('update_woodpecker') }}
        <a
          :href="`https://github.com/woodpecker-ci/woodpecker/releases/tag/${version.latest}`"
          target="_blank"
          rel="noopener noreferrer"
          class="underline"
          >{{ version.latest }}</a
        >
      </span>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import WoodpeckerLogo from '~/assets/logo.svg?component';
import Settings from '~/components/layout/Settings.vue';
import { useVersion } from '~/compositions/useVersion';

const version = useVersion();
</script>
