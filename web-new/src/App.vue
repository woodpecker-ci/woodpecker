<template>
  <div class="app flex flex-col m-auto w-full h-full">
    <Navbar />
    <router-view />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import Navbar from './components/Navbar.vue';

export default defineComponent({
  name: 'App',

  components: {
    Navbar,
  },
});
</script>

<style>
html,
body,
#app {
  width: 100%;
  height: 100%;
}
</style>

<style scoped>
.app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
