import '~/style/prism.css';

import Prism from 'prismjs';
import { computed, defineComponent, h, toRef, VNode } from 'vue';

declare type Data = Record<string, unknown>;

export default defineComponent({
  name: 'SyntaxHighlight',

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
    const code = toRef(props, 'code');
    const language = toRef(props, 'language');
    const prismLanguage = computed(() => Prism.languages[language.value]);
    const className = computed(() => `language-${language.value}`);

    return (): VNode =>
      h('pre', { ...attrs, class: [attrs.class, className] }, [
        h('code', {
          class: className,
          innerHTML: Prism.highlight(code.value, prismLanguage.value, language.value),
        }),
      ]);
  },
});
