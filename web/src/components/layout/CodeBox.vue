<template>
  <div class="code-box"><slot /></div>
</template>

<style scoped>
.code-box {
  @apply bg-gray-500 p-2 rounded-md text-white break-words dark:bg-dark-400 dark:text-gray-400;
  white-space: pre-wrap;
}
</style>
