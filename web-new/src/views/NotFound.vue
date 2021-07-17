<template>
  <div class="not-found">
    <p>Whoa 404, either we broke something or you had a typing mishap :-/</p>
    <router-link :to="{ name: 'home' }">Back to home</router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'NotFound',
});
</script>
