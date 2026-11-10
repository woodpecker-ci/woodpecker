import { useI18n } from 'vue-i18n';

let currentLocale = 'en';

function splitDuration(durationMs: number) {
  const totalSeconds = durationMs / 1000;
  const totalMinutes = totalSeconds / 60;
  const totalHours = totalMinutes / 60;

  const seconds = Math.floor(totalSeconds) % 60;
  const minutes = Math.floor(totalMinutes) % 60;
  const hours = Math.floor(totalHours) % 24;

  return {
    seconds,
    minutes,
    hours,
    totalHours,
    totalMinutes,
    totalSeconds,
  };
}

function toLocaleString(date: Date, tz?: string) {
  return date.toLocaleString(currentLocale, {
    dateStyle: 'short',
    timeStyle: 'short',
    timeZone: tz,
  });
}

function timeAgo(date: number) {
  const seconds = Math.floor((Date.now() - date) / 1000);

  const formatter = new Intl.RelativeTimeFormat(currentLocale);

  let interval = seconds / 31536000;
  if (interval > 1) {
    return formatter.format(-Math.round(interval), 'year');
  }
  interval = seconds / 2592000;
  if (interval > 1) {
    return formatter.format(-Math.round(interval), 'month');
  }
  interval = seconds / 86400;
  if (interval > 1) {
    return formatter.format(-Math.round(interval), 'day');
  }
  interval = seconds / 3600;
  if (interval > 1) {
    return formatter.format(-Math.round(interval), 'hour');
  }
  interval = seconds / 60;
  if (interval > 0.5) {
    return formatter.format(-Math.round(interval), 'minute');
  }
  return useI18n().t('time.just_now');
}

function prettyDuration(durationMs: number) {
  const formatter = new Intl.DurationFormat(currentLocale, { style: 'long' });

  const totalSeconds = Math.round(durationMs / 1000);
  if (totalSeconds < 60) {
    // secondsDisplay keeps "0 seconds" instead of an empty string.
    return new Intl.DurationFormat(currentLocale, { style: 'long', secondsDisplay: 'always' }).format({
      seconds: totalSeconds,
    });
  }

  const totalMinutes = Math.round(durationMs / (60 * 1000));
  if (totalMinutes < 60) {
    return formatter.format({ minutes: totalMinutes });
  }

  // From one hour on, show hours and minutes, with minutes rounded to 5.
  const roundedMinutes = Math.round(durationMs / (5 * 60 * 1000)) * 5;
  return formatter.format({ hours: Math.floor(roundedMinutes / 60), minutes: roundedMinutes % 60 });
}

function durationAsNumber(durationMs: number): string {
  const { seconds, minutes, hours } = splitDuration(durationMs);

  const minSecFormat = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;

  if (hours > 0) {
    return `${hours.toString().padStart(2, '0')}:${minSecFormat}`;
  }

  return minSecFormat;
}

export function useDate() {
  async function setDateLocale(locale: string) {
    currentLocale = locale;
  }

  return {
    toLocaleString,
    timeAgo,
    prettyDuration,
    setDateLocale,
    durationAsNumber,
  };
}
