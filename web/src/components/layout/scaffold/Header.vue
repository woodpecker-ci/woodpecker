<template>
  <div
    class="items-center justify-between py-4"
    :class="{
      'grid grid-cols-[1fr,auto,1fr]': threeColumn,
      'flex flex-wrap': !threeColumn,
    }"
  >
    <div class="flex flex-wrap items-center justify-start">
      <IconButton v-if="goBack" icon="back" :title="$t('back')" class="mr-2" @click="goBack" />
      <h1 class="flex text-xl text-color items-center gap-x-2">
        <slot name="title" />
      </h1>
    </div>
    <div v-if="threeColumn" class="flex items-center justify-center gap-x-2">
      <slot name="centerBox" />
    </div>
    <div class="flex items-center justify-end gap-x-2">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps({
  goBack: {
    type: Function,
    default: null,
  },
  threeColumn: {
    type: Boolean,
  },
});
</script>
