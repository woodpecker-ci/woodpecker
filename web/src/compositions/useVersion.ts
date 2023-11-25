import { onMounted, ref } from 'vue';

import useAuthentication from './useAuthentication';
import useConfig from './useConfig';

type VersionInfo = {
  latest: string;
  rc: string;
  next: string;
};

const version = ref<{
  latest: string | undefined;
  current: string;
  currentShort: string;
  needsUpdate: boolean;
  usesNext: boolean;
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
  const usesNext = current.startsWith('next');

  const { user } = useAuthentication();
  if (!user?.admin) {
    version.value = {
      latest: undefined,
      current,
      currentShort: usesNext ? 'next' : current,
      needsUpdate: false,
      usesNext,
    };
    return version;
  }

  if (current === 'dev') {
    version.value = {
      latest: undefined,
      current,
      currentShort: current,
      needsUpdate: false,
      usesNext,
    };
    return version;
  }

  onMounted(async () => {
    const versionInfo = await fetchVersion();

    let latest = undefined;
    if (versionInfo) {
      if (usesNext) {
        latest = versionInfo.next;
      } else if (current.includes('rc')) {
        latest = versionInfo.rc;
      } else {
        latest = versionInfo.latest;
      }
    }

    version.value = {
      latest,
      current,
      currentShort: usesNext ? 'next' : current,
      needsUpdate: latest !== undefined && latest !== current,
      usesNext,
    };
  });

  return version;
}
