import 'dayjs/locale/en';

import dayjs from 'dayjs';
import advancedFormat from 'dayjs/plugin/advancedFormat';
import timezone from 'dayjs/plugin/timezone';
import utc from 'dayjs/plugin/utc';
import { useI18n } from 'vue-i18n';

dayjs.extend(timezone);
dayjs.extend(utc);
dayjs.extend(advancedFormat);
dayjs.locale(navigator.language.split('-')[0]);

export function useDate() {
  function toLocaleString(date: Date) {
    return dayjs(date).format(useI18n().t('time.tmpl'));
  }

  return {
    toLocaleString,
  };
}
