import { onBeforeUnmount, onMounted, ref } from 'vue';

export function useInterval(fn: () => void | Promise<void>, ms: number): void {
  const id = ref<number>();

  onMounted(async () => {
    await fn(); // run once immediately
    id.value = window.setInterval(() => {
      void fn();
    }, ms);
  });

  onBeforeUnmount(() => {
    if (id.value != null) {
      window.clearInterval(id.value);
    }
  });
}
