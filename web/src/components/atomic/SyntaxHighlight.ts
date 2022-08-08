import '~/style/prism.css';

import Prism from 'prismjs';
import * as Vue from 'vue';
import { VNode } from 'vue';

declare type Data = Record<string, unknown>;

export default Vue.defineComponent({
  props: {
    code: {
      type: String,
      default: '',
    },
    language: {
      type: String,
      default: 'yaml',
    },
  },
  setup(props, { attrs }: { attrs: Data }) {
    const { h } = Vue;
    const { code, language } = props;
    const prismLanguage = Prism.languages[language];
    const className = `language-${language}`;

    return (): VNode =>
      h('pre', { ...attrs, class: [attrs.class, className] }, [
        h('code', {
          class: className,
          innerHTML: Prism.highlight(code, prismLanguage, language),
        }),
      ]);
  },
});
