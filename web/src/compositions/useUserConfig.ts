import { useStorage } from '@vueuse/core';
import { computed } from 'vue';

interface UserConfig {
  isPipelineFeedOpen: boolean;
  redirectUrl: string;
  collapseLogGroupsByDefault: boolean;
}

const config = useStorage<UserConfig>(
  'woodpecker:user-config',
  {
    isPipelineFeedOpen: false,
    redirectUrl: '',
    collapseLogGroupsByDefault: true,
  },
  undefined,
  { mergeDefaults: true },
);

export default () => ({
  setUserConfig<T extends keyof UserConfig>(key: T, value: UserConfig[T]): void {
    config.value = { ...config.value, [key]: value };
  },
  userConfig: computed(() => config.value),
});
