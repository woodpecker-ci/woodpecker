import { computed, ref, watch } from 'vue';

import useConfig from '~/compositions/useConfig';
import { useTheme } from '~/compositions/useTheme';
import { PipelineStatus } from '~/lib/api/types';

const { resolvedTheme } = useTheme();
const darkMode = computed(() => resolvedTheme.value);

type Status = 'default' | 'success' | 'pending' | 'error';
const faviconStatus = ref<Status>('default');

watch(
  [darkMode, faviconStatus],
  () => {
    const faviconPNG = document.getElementById('favicon-png');
    if (faviconPNG) {
      (faviconPNG as HTMLLinkElement).href = `${useConfig().rootPath}/favicons/favicon-${darkMode.value}-${
        faviconStatus.value
      }.png`;
    }

    const faviconSVG = document.getElementById('favicon-svg');
    if (faviconSVG) {
      (faviconSVG as HTMLLinkElement).href = `${useConfig().rootPath}/favicons/favicon-${darkMode.value}-${
        faviconStatus.value
      }.svg`;
    }
  },
  { immediate: true },
);

function convertStatus(status: PipelineStatus): Status {
  if (['blocked', 'declined', 'error', 'failure', 'killed'].includes(status)) {
    return 'error';
  }

  if (['started', 'running', 'pending'].includes(status)) {
    return 'pending';
  }

  if (['success', 'declined', 'error', 'failure', 'killed'].includes(status)) {
    return 'success';
  }

  // skipped
  return 'default';
}

export function useFavicon() {
  return {
    updateStatus(status?: PipelineStatus | 'default') {
      if (status === undefined || status === 'default') {
        faviconStatus.value = 'default';
      } else {
        faviconStatus.value = convertStatus(status);
      }
    },
  };
}
