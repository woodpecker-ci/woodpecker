import { onBeforeUnmount, onMounted, ref } from 'vue';

export function useInterval(fn: () => void | Promise<void>, ms: number): void {
  const id = ref<number>();
  let unmounted = false;

  onMounted(async () => {
    await fn(); // run once immediately
    if (unmounted) {
      // component was unmounted while the first call was in flight,
      // starting the interval now would leak it forever
      return;
    }
    id.value = window.setInterval(() => {
      void fn();
    }, ms);
  });

  onBeforeUnmount(() => {
    unmounted = true;
    if (id.value != null) {
      window.clearInterval(id.value);
    }
  });
}
