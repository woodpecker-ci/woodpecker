import semverCoerce from 'semver/functions/coerce';
import semverGt from 'semver/functions/gt';
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

const isInitialized = ref(false);

export function useVersion() {
  if (isInitialized.value) {
    return version;
  }
  isInitialized.value = true;

  const config = useConfig();
  const current = config.version as string;
  const currentSemver = semverCoerce(current);
  const usesNext = current.startsWith('next');

  const { user } = useAuthentication();
  if (config.skipVersionCheck || !user?.admin) {
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

    let latest;
    if (versionInfo) {
      if (usesNext) {
        latest = versionInfo.next;
      } else if (current.includes('rc')) {
        latest = versionInfo.rc;
      } else {
        latest = versionInfo.latest;
      }
    }

    let needsUpdate = false;
    if (usesNext) {
      needsUpdate = latest !== current;
    } else if (latest !== undefined && currentSemver !== null) {
      needsUpdate = semverGt(latest, currentSemver);
    }

    version.value = {
      latest,
      current,
      currentShort: usesNext ? 'next' : current,
      needsUpdate,
      usesNext,
    };
  });

  return version;
}
