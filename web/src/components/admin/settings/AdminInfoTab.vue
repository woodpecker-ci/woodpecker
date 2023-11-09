<template>
  <Settings :title="$t('info')">
    <div class="flex flex-col items-center gap-4">
      <WoodpeckerLogo class="w-32 h-32 fill-wp-text-200" />

      <i18n-t keypath="running_version" tag="p" class="text-xl">
        <span class="font-bold">{{ version?.current }}</span>
      </i18n-t>

      <i18n-t v-if="version?.needsUpdate" keypath="update_woodpecker" tag="span" class="text-int-wp-state-error-100">
        <a
          :href="`https://github.com/woodpecker-ci/woodpecker/releases/tag/${version.latest}`"
          target="_blank"
          rel="noopener noreferrer"
          class="underline"
          >{{ version.latest }}</a
        >
      </i18n-t>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import WoodpeckerLogo from '~/assets/logo.svg?component';
import Settings from '~/components/layout/Settings.vue';
import { useVersion } from '~/compositions/useVersion';

const version = useVersion();
</script>
