import { onBeforeUnmount, onMounted, ref } from 'vue';

export function useInterval(fn: () => void | Promise<void>, ms: number, options?: { immediate?: boolean }): void {
  const id = ref<number | null>(null);

  onMounted(async () => {
    if ((options?.immediate ?? true) === true) {
      await fn(); // run once immediately
    }
    id.value = window.setInterval(() => {
      void fn();
    }, ms);
  });

  onBeforeUnmount(() => {
    if (id.value !== null) {
      window.clearInterval(id.value);
    }
  });
}
