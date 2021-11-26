<template>
  <div class="flex flex-col h-full w-full items-center justify-center">
    <p class="text-2xl mb-8">Whoa 404, either we broke something or you had a typing mishap :-/</p>
    <span>Back to <router-link class="text-blue-400" replace :to="{ name: 'home' }">home</router-link></span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'NotFound',
});
</script>
