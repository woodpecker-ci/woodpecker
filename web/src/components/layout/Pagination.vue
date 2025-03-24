<template>
  <div>
    <div v-if="loading">
      <Icon name="loading" class="animate-spin" />
    </div>
    <div v-else>
      <slot />
    </div>

    <Button v-if="hasMore" :is-loading="loading" :text="$t('load_more')" class="mx-auto mt-4" @click="nextPage" />
  </div>
</template>

<script lang="ts" setup>
import { computed, nextTick } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import type { usePagination } from '~/compositions/usePaginate';

const props = defineProps<{
  pagination: ReturnType<typeof usePagination>;
}>();
const loading = computed(() => props.pagination.loading.value);
const hasMore = computed(() => props.pagination.hasMore.value);
function scrollDown() {
  window.scrollTo({
    top: document.body.scrollHeight,
    behavior: 'smooth',
  });
}
function nextPage() {
  props.pagination.nextPage();
  nextTick(() => {
    setTimeout(scrollDown, 100);
  });
}
</script>
