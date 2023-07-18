<template>
  <div class="flex flex-col h-full w-full items-center justify-center">
    <p class="text-2xl mb-8">{{ $t('not_found.not_found') }}</p>
    <span
      ><router-link class="text-wp-blue-400" replace :to="{ name: 'home' }">{{
        $t('not_found.back_home')
      }}</router-link></span
    >
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'NotFound',
});
</script>
