import { useTitle } from '@vueuse/core';
import type { Ref } from 'vue';
import { computed } from 'vue';

export function useWPTitle(elements: Ref<string[]>) {
  useTitle(computed(() => `${elements.value.join(' · ')} · Woodpecker`));
}
