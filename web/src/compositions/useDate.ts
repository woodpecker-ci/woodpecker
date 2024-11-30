import type { LocaleData } from 'javascript-time-ago';
import TimeAgo from 'javascript-time-ago';
import en from 'javascript-time-ago/locale/en';

TimeAgo.addDefaultLocale(en);
let ta = new TimeAgo('en');
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

function toLocaleString(date: Date) {
  return date.toLocaleString('de', {
    dateStyle: 'short',
    timeStyle: 'short',
  });
}

function timeAgo(date: Date) {
  return ta.format(date);
}

function prettyDuration(durationMs: number) {
  const t = splitDuration(durationMs);

  if (t.totalHours > 1) {
    return Intl.NumberFormat(currentLocale, { style: 'unit', unit: 'hour', unitDisplay: 'long' }).format(
      Math.round(t.totalHours),
    );
  }
  if (t.totalMinutes > 1) {
    return Intl.NumberFormat(currentLocale, { style: 'unit', unit: 'minute', unitDisplay: 'long' }).format(
      Math.round(t.totalMinutes),
    );
  }
  return Intl.NumberFormat(currentLocale, { style: 'unit', unit: 'second', unitDisplay: 'long' }).format(
    Math.round(t.totalSeconds),
  );
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
  const addedLocales = ['en'];

  async function setDayjsLocale(locale: string) {
    currentLocale = locale;
    if (!addedLocales.includes(locale)) {
      const l = (await import(`~/assets/timeAgoLocales/${locale}.json`)) as LocaleData;
      TimeAgo.addLocale(l);
    }
    ta = new TimeAgo(locale);
  }

  return {
    toLocaleString,
    timeAgo,
    prettyDuration,
    setDayjsLocale,
    durationAsNumber,
  };
}
