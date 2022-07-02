import 'dayjs/locale/en';
import 'dayjs/locale/lv';
import 'dayjs/locale/de';

import dayjs from 'dayjs';
import advancedFormat from 'dayjs/plugin/advancedFormat';
import timezone from 'dayjs/plugin/timezone';
import utc from 'dayjs/plugin/utc';
import { useI18n } from 'vue-i18n';

import { getUserLanguage } from '~/utils/locale';

dayjs.extend(timezone);
dayjs.extend(utc);
dayjs.extend(advancedFormat);
dayjs.locale(getUserLanguage());

export function useDate() {
  function toLocaleString(date: Date) {
    return dayjs(date).format(useI18n().t('time.tmpl'));
  }

  return {
    toLocaleString,
  };
}
