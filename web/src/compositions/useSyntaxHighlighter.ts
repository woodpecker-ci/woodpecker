import { createHighlighter, type BundledLanguage, type BundledTheme, type HighlighterGeneric } from 'shiki';
import { computed, onMounted, Ref, ref, watch } from 'vue';

import { useTheme } from '~/compositions/useTheme';

const highlighter = ref<HighlighterGeneric<BundledLanguage, BundledTheme>>();
const isLoading = ref(false);

export default function useSyntaxHighlighter(code: Ref<string>, language: Ref<BundledLanguage> = ref('yaml')) {
  const { theme: currentTheme } = useTheme();
  const shikiTheme = computed(() => (currentTheme.value === 'dark' ? 'github-dark' : 'github-light'));

  onMounted(async () => {
    // only the first call should load the highlighter
    if (highlighter.value || isLoading.value) return;

    isLoading.value = true;
    highlighter.value = await createHighlighter({
      themes: [shikiTheme.value],
      langs: ['yaml'],
    });
    isLoading.value = false;
  });

  const formattedCode = ref<string>();

  watch(
    [code, shikiTheme, highlighter],
    async () => {
      if (!highlighter.value) return;

      await Promise.all([
        highlighter.value.loadTheme(shikiTheme.value),
        highlighter.value.loadLanguage(language.value ?? 'yaml'),
      ]);

      formattedCode.value = highlighter.value.codeToHtml(code.value, {
        lang: language.value ?? 'yaml',
        theme: shikiTheme.value,
        transformers: [
          {
            preprocess(code) {
              // Workaround for https://github.com/shikijs/shiki/issues/608
              // When last span is empty, it's height is 0px
              // so add a newline to render it correctly
              if (code.endsWith('\n')) return `${code}\n`;
            },
          },
        ],
      });
    },
    { immediate: true },
  );

  return { formattedCode, isLoading };
}
