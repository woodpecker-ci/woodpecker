import type { MaybeRef } from 'vue';
import { computed, ref, unref, watch } from 'vue';

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

export async function useAsyncData<A extends unknown[], R>(
  action: (...a: A) => Promise<R>,
  args: { [K in keyof A]: MaybeRef<A[K]> },
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
      const unwrappedArgs = args.map((a) => unref(a)) as A;
      data.value = await action(...unwrappedArgs);
    } catch (_error) {
      console.error(_error);
      options.onError?.(_error);
      error.value = _error;
    } finally {
      isLoading.value = false;
    }
  }

  watch(args, doFetch);

  if (options.immediate) {
    await doFetch();
  }

  return {
    refetch: doFetch,
    isLoading: computed(() => isLoading.value),
    data: computed(() => data.value as R | null),
    error: computed(() => error.value),
  };
}
