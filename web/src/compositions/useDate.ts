import dayjs from 'dayjs';
import advancedFormat from 'dayjs/plugin/advancedFormat';
import timezone from 'dayjs/plugin/timezone';
import utc from 'dayjs/plugin/utc';

dayjs.extend(timezone);
dayjs.extend(utc);
dayjs.extend(advancedFormat);

export function useDate() {
  function toLocaleString(date: Date) {
    return dayjs(date).format('MMM D, YYYY, HH:mm z');
  }

  return {
    toLocaleString,
  };
}
