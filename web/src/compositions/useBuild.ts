import { computed, Ref } from 'vue';

import { useElapsedTime } from '~/compositions/useElapsedTime';
import { Build } from '~/lib/api/types';
import { prettyDuration } from '~/utils/duration';
import { convertEmojis } from '~/utils/emoji';
import timeAgo from '~/utils/timeAgo';

export default (build: Ref<Build | undefined>) => {
  const sinceRaw = computed(() => {
    if (!build.value) {
      return undefined;
    }

    const start = build.value.created_at || 0;

    return start * 1000;
  });

  const sinceUnderOneHour = computed(
    () => sinceRaw.value !== undefined && sinceRaw.value > 0 && sinceRaw.value <= 1000 * 60 * 60,
  );
  const { time: sinceElapsed } = useElapsedTime(sinceUnderOneHour, sinceRaw);

  const since = computed(() => {
    if (sinceRaw.value === 0) {
      return 'not started yet';
    }

    if (sinceElapsed.value === undefined) {
      return null;
    }

    return timeAgo.format(sinceElapsed.value);
  });

  const durationRaw = computed(() => {
    if (!build.value) {
      return undefined;
    }

    const start = build.value.started_at || 0;
    const end = build.value.finished_at || 0;

    if (start === 0) {
      return 0;
    }

    if (end === 0) {
      return Date.now() - start * 1000;
    }

    return (end - start) * 1000;
  });

  const running = computed(() => build.value !== undefined && build.value.status === 'running');
  const { time: durationElapsed } = useElapsedTime(running, durationRaw);

  const duration = computed(() => {
    if (durationElapsed.value === undefined) {
      return null;
    }

    if (durationRaw.value === 0) {
      return 'not started yet';
    }

    return prettyDuration(durationElapsed.value);
  });

  const message = computed(() => {
    if (!build.value) {
      return '';
    }

    return convertEmojis(build.value.message);
  });

  const prettyRef = computed(() => {
    if (build.value?.event === 'push') {
      return build.value.branch;
    }

    if (build.value?.event === 'tag') {
      return build.value.ref.replaceAll('refs/tags/', '');
    }

    if (build.value?.event === 'pull_request') {
      return `#${build.value.ref
        .replaceAll('refs/pull/', '')
        .replaceAll('refs/merge-requests/', '')
        .replaceAll('/merge', '')
        .replaceAll('/head', '')}`;
    }

    return build.value?.ref;
  });

  return { since, duration, message, prettyRef };
};
