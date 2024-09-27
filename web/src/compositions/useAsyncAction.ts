import { computed, ref } from 'vue';

export function useAsyncAction<T extends unknown[]>(
  action: (...a: T) => void | Promise<void>,
  onerror: ((error: any) => void) | undefined = undefined,
) {
  const isLoading = ref(false);

  async function doSubmit(...a: T) {
    if (isLoading.value) {
      return;
    }

    isLoading.value = true;
    try {
      await action(...a);
    } catch (error) {
      console.error(error);
      if (onerror) {
        onerror(error);
      }
    }
    isLoading.value = false;
  }

  return {
    doSubmit,
    isLoading: computed(() => isLoading.value),
  };
}
