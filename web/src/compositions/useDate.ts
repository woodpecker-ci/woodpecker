import dayjs from 'dayjs';
import advancedFormat from 'dayjs/plugin/advancedFormat';
import timezone from 'dayjs/plugin/timezone';
import utc from 'dayjs/plugin/utc';
import relativeTime from 'dayjs/plugin/relativeTime'
import { useI18n } from 'vue-i18n';
import duration from 'dayjs/plugin/duration';

dayjs.extend(timezone);
dayjs.extend(utc);
dayjs.extend(advancedFormat);
dayjs.extend(relativeTime);
dayjs.extend(duration);

// TODO improve
window.dayjs = dayjs;

export function useDate() {
  function toLocaleString(date: Date) {
    return dayjs(date).format(useI18n().t('time.tmpl'));
  }

  function timeAgo(date: Date|string|number) {
    return dayjs().to(dayjs(date))
  }

  function prettyDuration(duration: number) {
    return dayjs.duration(duration).humanize()
  }

  const addedLocales = ['en'];

  async function setDayjsLocale(locale: string) {
    if (!addedLocales.includes(locale)) {
      await import(`~/assets/dayjsLocales/${locale}.js`);
    }
    dayjs.locale(locale);
  }

  return {
    toLocaleString,
    timeAgo,
    prettyDuration,
    setDayjsLocale,
  };
}
