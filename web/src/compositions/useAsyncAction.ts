import type { MaybeRef } from 'vue';
import { computed, isRef, ref, unref, watch } from 'vue';

export function useAsyncAction<T extends unknown[]>(
  action: (...a: T) => void | Promise<void>,
  onerror: ((error: any) => void) | undefined = undefined,
) {
  const isLoading = ref(false);
  const error = ref<unknown>(null);

  async function doSubmit(...a: T) {
    if (isLoading.value) {
      return;
    }

    isLoading.value = true;
    try {
      await action(...a);
    } catch (_error) {
      console.error(_error);
      onerror?.(_error);
      error.value = _error;
    }
    isLoading.value = false;
  }

  return {
    doSubmit,
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}

export function useAsyncData<R>(
  action: MaybeRef<() => Promise<R>>,
  options: { immediate?: boolean; onError?: (error: unknown) => void } = { immediate: true },
) {
  const isLoading = ref(false);
  const data = ref<R | null>(null);
  const error = ref<unknown>(null);

  async function doFetch() {
    if (isLoading.value) {
      return;
    }

    isLoading.value = true;
    error.value = null;
    try {
      data.value = await unref(action)();
    } catch (_error) {
      console.error(_error);
      options.onError?.(_error);
      error.value = _error;
    } finally {
      isLoading.value = false;
    }
  }

  if (isRef(action)) {
    watch(action, doFetch);
  }

  if (options.immediate) {
    void doFetch();
  }

  return {
    refetch: doFetch,
    isLoading: computed(() => isLoading.value),
    data: computed(() => data.value as R | null),
    error: computed(() => error.value),
  };
}
