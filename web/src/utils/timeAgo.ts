import TimeAgo from 'javascript-time-ago';
import en from 'javascript-time-ago/locale/en.json';
import lv from 'javascript-time-ago/locale/lv.json';

import { getUserLanguage } from '~/utils/locale';

TimeAgo.addDefaultLocale(en);
TimeAgo.addLocale(lv);

const timeAgo = new TimeAgo(getUserLanguage());

export default timeAgo;
