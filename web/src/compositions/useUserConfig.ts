import { useStorage } from '@vueuse/core';
import { computed } from 'vue';

type UserConfig = {
  isPipelineFeedOpen: boolean;
  redirectUrl: string;
};

const config = useStorage<UserConfig>('woodpecker:user-config', {
  isPipelineFeedOpen: false,
  redirectUrl: '',
});

export default () => ({
  setUserConfig<T extends keyof UserConfig>(key: T, value: UserConfig[T]): void {
    config.value = { ...config.value, [key]: value };
  },
  userConfig: computed(() => config.value),
});
