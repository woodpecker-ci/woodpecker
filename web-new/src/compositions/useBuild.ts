import { computed, Ref } from 'vue';
import { Build } from '~/lib/api/types';
import timeAgo from '~/utils/timeAgo';
import { convertEmojis } from '~/utils/emoji';
import { prettyDuration } from '~/utils/duration';

export default (build: Ref<Build | undefined>) => {
  const since = computed(() => {
    if (!build.value) {
      return null;
    }

    const start = build.value.started_at || 0;

    if (start === 0) {
      return 'not started yet';
    }

    return timeAgo.format(start * 1000);
  });

  const duration = computed(() => {
    if (!build.value) {
      return null;
    }

    const start = build.value.started_at || 0;
    const end = build.value.finished_at || 0;

    if (start === 0) {
      return 'not started yet';
    }

    if (end === 0) {
      return prettyDuration(Date.now() - start * 1000);
    }

    return prettyDuration((end - start) * 1000);
  });

  const message = computed(() => {
    if (!build.value) {
      return '';
    }

    return convertEmojis(build.value.message);
  });

  return { since, duration, message };
};
