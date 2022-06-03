import TimeAgo from 'javascript-time-ago';
import en from 'javascript-time-ago/locale/en.json';

TimeAgo.addDefaultLocale(en);

const timeAgo = new TimeAgo(navigator.language);

export default timeAgo;
