import { computed, ref } from 'vue';

const USER_CONFIG_KEY = 'woodpecker-user-config';

type UserConfig = {
  isBuildFeedOpen: boolean;
  redirectUrl: string;
};

const defaultUserConfig: UserConfig = {
  isBuildFeedOpen: false,
  redirectUrl: '',
};

function loadUserConfig(): UserConfig {
  const lsData = localStorage.getItem(USER_CONFIG_KEY);
  if (!lsData) {
    return defaultUserConfig;
  }

  return JSON.parse(lsData);
}

const config = ref<UserConfig>(loadUserConfig());

export default () => ({
  setUserConfig<T extends keyof UserConfig>(key: T, value: UserConfig[T]): void {
    config.value = { ...config.value, [key]: value };
    localStorage.setItem(USER_CONFIG_KEY, JSON.stringify(config.value));
  },
  userConfig: computed(() => config.value),
});
