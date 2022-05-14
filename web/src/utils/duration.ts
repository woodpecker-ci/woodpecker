import humanizeDuration from 'humanize-duration';
import { useI18n } from 'vue-i18n';

export function prettyDuration(durationMs: number): string {
  const i18n = useI18n();
  const short = {
    w: () => i18n.t('time.weeks_short'),
    d: () => i18n.t('time.days_short'),
    h: () => i18n.t('time.hours_short'),
    m: () => i18n.t('time.min_short'),
    s: () => i18n.t('time.sec_short'),
  };
  const durationOptions: humanizeDuration.HumanizerOptions = {
    round: true,
    languages: { short },
    language: 'short',
  };

  if (durationMs < 1000 * 60 * 60) {
    return humanizeDuration(durationMs, durationOptions);
  }
  return humanizeDuration(durationMs, { ...durationOptions, units: ['y', 'mo', 'd', 'h', 'm'] });
}

function leadingZeros(n: number, length: number): string {
  let res = n.toString();
  while (res.length < length) {
    res = `0${res}`;
  }
  return res;
}

export function durationAsNumber(durationMs: number): string {
  const durationSeconds = durationMs / 1000;
  const seconds = leadingZeros(Math.floor(durationSeconds % 60), 2);
  const minutes = leadingZeros(Math.floor(durationSeconds / 60) % 60, 2);
  const hours = Math.floor(durationSeconds / 3600);

  if (hours !== 0) {
    return `${hours}:${minutes}:${seconds}`;
  }

  return `${minutes}:${seconds}`;
}
