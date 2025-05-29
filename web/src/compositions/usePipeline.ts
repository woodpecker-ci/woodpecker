import { emojify } from 'node-emoji';
import { computed } from 'vue';
import type { Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { useDate } from '~/compositions/useDate';
import { useElapsedTime } from '~/compositions/useElapsedTime';
import type { Pipeline } from '~/lib/api/types';

const { toLocaleString, timeAgo, prettyDuration } = useDate();

export default (pipeline: Ref<Pipeline | undefined>) => {
  const sinceRaw = computed(() => {
    if (!pipeline.value) {
      return undefined;
    }

    const start = pipeline.value.created || 0;

    return start * 1000;
  });

  const sinceUnderOneHour = computed(
    () => sinceRaw.value !== undefined && sinceRaw.value > 0 && sinceRaw.value <= 1000 * 60 * 60,
  );
  const { time: sinceElapsed } = useElapsedTime(sinceUnderOneHour, sinceRaw);

  const i18n = useI18n();
  const since = computed(() => {
    if (sinceRaw.value === 0) {
      // return i18n.t('time.not_started');
      return '-';
    }

    if (sinceElapsed.value === undefined) {
      return null;
    }

    // TODO: check whether elapsed works
    return timeAgo(sinceElapsed.value);
  });

  const durationRaw = computed(() => {
    if (!pipeline.value) {
      return undefined;
    }

    const start = pipeline.value.started || 0;
    const end = pipeline.value.finished || pipeline.value.updated || 0;

    if (start === 0 || end === 0) {
      return 0;
    }

    // only calculate time based no now() for running pipelines
    if (pipeline.value.status === 'running') {
      return Date.now() - start * 1000;
    }

    return (end - start) * 1000;
  });

  const running = computed(() => pipeline.value !== undefined && pipeline.value.status === 'running');
  const { time: durationElapsed } = useElapsedTime(running, durationRaw);

  const duration = computed(() => {
    if (durationElapsed.value === undefined) {
      return null;
    }

    if (durationRaw.value === 0) {
      return i18n.t('time.not_started');
    }

    return prettyDuration(durationElapsed.value);
  });

  const message = computed(() => emojify(pipeline.value?.commit_pipeline.message ?? ''));
  const shortMessage = computed(() => message.value.split('\n')[0]);

  const context = computed(() => {
    let context = '';
    if (pipeline.value?.event === 'pull_request' || pipeline.value?.event === 'pull_request_closed') {
      context = pipeline.value.pull_request!.title;
    } else if (pipeline.value?.event === 'deployment') {
      context = pipeline.value.deployment!.description;
    } else if (pipeline.value?.event === 'release' || pipeline.value?.event === 'tag') {
      context = pipeline.value.release_tag_title || '';
    }
    return emojify(context);
  });
  const shortContext = computed(() => context.value.split('\n')[0]);

  const prettyRef = computed(() => {
    if (pipeline.value?.event === 'tag' || pipeline.value?.event === 'release') {
      return pipeline.value.ref.replaceAll('refs/tags/', '');
    }

    if (pipeline.value?.event === 'pull_request' || pipeline.value?.event === 'pull_request_closed') {
      return `#${pipeline.value.pull_request!.index}`;
    }

    if (!pipeline.value) {
      return '';
    }

    return pipeline.value.branch || pipeline.value.ref;
  });

  const created = computed(() => {
    if (!pipeline.value) {
      return undefined;
    }

    const start = pipeline.value.created || 0;

    return toLocaleString(new Date(start * 1000));
  });

  return { since, duration, durationElapsed, message, shortMessage, shortContext, context, prettyRef, created };
};
