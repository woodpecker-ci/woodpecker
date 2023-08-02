import TimeAgo from 'javascript-time-ago';
import de from 'javascript-time-ago/locale/de.json';
import en from 'javascript-time-ago/locale/en.json';
import es from 'javascript-time-ago/locale/es.json';
import fi from 'javascript-time-ago/locale/fi.json';
import fr from 'javascript-time-ago/locale/fr.json';
import hu from 'javascript-time-ago/locale/hu.json';
import id from 'javascript-time-ago/locale/id.json';
import it from 'javascript-time-ago/locale/it.json';
import lv from 'javascript-time-ago/locale/lv.json';
import nl from 'javascript-time-ago/locale/nl.json';
import pl from 'javascript-time-ago/locale/pl.json';
import pt from 'javascript-time-ago/locale/pt.json';
import ru from 'javascript-time-ago/locale/ru.json';
import uk from 'javascript-time-ago/locale/uk.json';
import zhHans from 'javascript-time-ago/locale/zh.json'; // FIXME which zh-Hans?
import zhHant from 'javascript-time-ago/locale/zh-Hant.json';

import { getUserLanguage } from '~/utils/locale';

TimeAgo.addDefaultLocale(en);
TimeAgo.addLocale(de);
TimeAgo.addLocale(es);
TimeAgo.addLocale(fi);
TimeAgo.addLocale(fr);
TimeAgo.addLocale(hu);
TimeAgo.addLocale(id);
TimeAgo.addLocale(it);
TimeAgo.addLocale(lv);
TimeAgo.addLocale(nl);
TimeAgo.addLocale(pl);
TimeAgo.addLocale(pt);
TimeAgo.addLocale(ru);
TimeAgo.addLocale(uk);
TimeAgo.addLocale(zhHans);
TimeAgo.addLocale(zhHant);

const timeAgo = new TimeAgo(getUserLanguage());

export default timeAgo;
