import { onMounted, ref } from 'vue';

import useAuthentication from './useAuthentication';
import useConfig from './useConfig';

type VersionInfo = {
  latest: string;
  next: string;
};

const version = ref<{
  latest: string | undefined;
  current: string;
  currentShort: string;
  needsUpdate: boolean;
}>();

async function fetchVersion(): Promise<VersionInfo | undefined> {
  try {
    const resp = await fetch('https://woodpecker-ci.org/version.json');
    const json = await resp.json();
    return json;
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Failed to fetch version info', error);
    return undefined;
  }
}

const isInitialised = ref(false);

export function useVersion() {
  if (isInitialised.value) {
    return version;
  }
  isInitialised.value = true;

  const config = useConfig();
  const current = config.version as string;
  const usesNext = config.version?.startsWith('next');

  const { user } = useAuthentication();
  if (!user?.admin) {
    version.value = {
      latest: undefined,
      current,
      currentShort: usesNext ? 'next' : current,
      needsUpdate: false,
    };
  }

  onMounted(async () => {
    const versionInfo = await fetchVersion();

    let needsUpdate = false;
    if (versionInfo) {
      if (current === 'dev') {
        needsUpdate = false;
      } else if (usesNext) {
        needsUpdate = versionInfo.next !== current;
      } else {
        needsUpdate = versionInfo.latest !== current;
      }
    }

    version.value = {
      latest: usesNext ? versionInfo?.next : versionInfo?.latest,
      current,
      currentShort: usesNext ? 'next' : current,
      needsUpdate,
    };
  });

  return version;
}
