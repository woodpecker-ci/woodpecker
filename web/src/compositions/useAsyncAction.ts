import { computed, ref } from 'vue';

import useNotifications from '~/compositions/useNotifications';

const notifications = useNotifications();

export type UseSubmitOptions = {
  showErrorNotification: false;
};

export function useAsyncAction<T extends unknown[]>(
  action: (...a: T) => void | Promise<void>,
  options?: UseSubmitOptions,
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
      if (options?.showErrorNotification) {
        notifications.notify({ title: (error as Error).message, type: 'error' });
      }
    }
    isLoading.value = false;
  }

  return {
    doSubmit,
    isLoading: computed(() => isLoading.value),
  };
}
