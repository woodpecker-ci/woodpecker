import dayjs from 'dayjs';
import advancedFormat from 'dayjs/plugin/advancedFormat';
import duration from 'dayjs/plugin/duration';
import relativeTime from 'dayjs/plugin/relativeTime';
import timezone from 'dayjs/plugin/timezone';
import utc from 'dayjs/plugin/utc';
import { useI18n } from 'vue-i18n';

dayjs.extend(timezone);
dayjs.extend(utc);
dayjs.extend(advancedFormat);
dayjs.extend(relativeTime);
dayjs.extend(duration);

function toLocaleString(date: Date) {
  return dayjs(date).format(useI18n().t('time.template'));
}

function timeAgo(date: Date | string | number) {
  return dayjs().to(dayjs(date));
}

function prettyDuration(durationMs: number) {
  return dayjs.duration(durationMs).humanize();
}

function durationAsNumber(durationMs: number): string {
  const dur = dayjs.duration(durationMs);
  return dur.format(dur.hours() > 1 ? 'HH:mm:ss' : 'mm:ss');
}

export function useDate() {
  const addedLocales = ['en'];

  async function setDayjsLocale(locale: string) {
    if (!addedLocales.includes(locale)) {
      const l = (await import(`~/assets/dayjsLocales/${locale}.js`)) as { default: string };
      dayjs.locale(l.default);
    } else {
      dayjs.locale(locale);
    }
  }

  return {
    toLocaleString,
    timeAgo,
    prettyDuration,
    setDayjsLocale,
    durationAsNumber,
  };
}
