import { computed, ref } from 'vue';

import useNotifications from '~/compositions/useNotifications';

const notifications = useNotifications();

export function useAsyncAction<T extends unknown[]>(action: (...a: T) => void | Promise<void>) {
  const isLoading = ref(false);

  async function doSubmit(...a: T) {
    if (isLoading.value) {
      return;
    }

    isLoading.value = true;
    try {
      await action(...a);
    } catch (error) {
      notifications.notify({ title: (error as Error).message, type: 'error' });
    }
    isLoading.value = false;
  }

  return {
    doSubmit,
    isLoading: computed(() => isLoading.value),
  };
}
