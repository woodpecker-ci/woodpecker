<template>
  <div
    class="items-center justify-between py-4"
    :class="{
      'grid grid-cols-[1fr,auto,1fr]': threeColumns,
      'flex flex-wrap': !threeColumns,
    }"
  >
    <div class="flex flex-wrap items-center justify-start">
      <IconButton v-if="goBack" icon="back" :title="$t('back')" class="mr-2" @click="goBack" />
      <h1 class="flex flex-wrap text-xl text-color items-center gap-x-2">
        <slot name="title" />
      </h1>
    </div>
    <div v-if="threeColumns" class="flex flex-wrap items-center justify-center gap-x-2">
      <slot name="centerBox" />
    </div>
    <div class="flex flex-wrap items-center justify-end gap-x-2">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSlots } from 'vue';

export interface Props {
  goBack?: () => void;
}

defineProps<Props>();

const slots = useSlots();

const threeColumns = !!slots.centerBox;
</script>
