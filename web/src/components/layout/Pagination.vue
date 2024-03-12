<template>
  <div>
    <div v-if="loading">loading ...</div>
    <div v-else>
      <slot :data="data" />
    </div>

    <Button v-if="hasMore" :is-loading="loading" text="Load more" class="mx-auto mt-4" @click="nextPage" />
  </div>
</template>

<script lang="ts" setup>
import { nextTick } from 'vue';

import Button from '~/components/atomic/Button.vue';

import { usePagination } from '~/compositions/usePaginate';

const props = defineProps<{
  loadData: () => Promise<unknown[] | null>;
}>();

const {
  loading,
  data,
  nextPage: _nextPage,
  hasMore,
} = usePagination(props.loadData, () => true, {
  name: 'secrets',
  each: ['repo', 'org', 'global'],
});

function scrollDown() {
  window.scrollTo({
    top: document.body.scrollHeight,
    behavior: 'smooth',
  });
}

function nextPage() {
  _nextPage();
  nextTick(() => {
    setTimeout(scrollDown, 100);
  });
}
</script>
