import { useInfiniteScroll } from '@vueuse/core';
import { onMounted, Ref, ref, watch, UnwrapRef } from 'vue';

export async function usePaginate<T>(getSingle: (page: number) => Promise<T[]>): Promise<T[]> {
  let hasMore = true;
  let page = 1;
  const result: T[] = [];
  while (hasMore) {
    // eslint-disable-next-line no-await-in-loop
    const singleRes = await getSingle(page);
    result.push(...singleRes);
    hasMore = singleRes.length !== 0;
    page += 1;
  }
  return result;
}

export function usePagination<T, S = unknown>(
  _loadData: (page: number, arg: S) => Promise<T[] | null>,
  isActive: () => boolean,
  {
    scrollElement: _scrollElement,
    name,
    each: _each,
  }: { scrollElement?: Ref<HTMLElement | null>; name?: string; each?: S[] } = {},
) {
  const scrollElement = _scrollElement ?? ref(document.getElementById('scroll-component'));
  const page = ref(1);
  const pageSize = ref(0);
  const hasMore = ref(true);
  const data = ref<T[]>([]) as Ref<T[]>;
  const loading = ref(false);
  const each = ref(_each ?? []);

  async function loadData() {
    if (loading.value === true || hasMore.value === false) {
      return;
    }

    loading.value = true;
    const newData = (await _loadData(page.value, each.value?.[0] as S)) ?? [];
    hasMore.value = newData.length >= pageSize.value && newData.length > 0;
    name === 'secrets' &&
      console.log(
        '>>',
        'load',
        each.value?.[0] as S,
        page.value,
        ':',
        newData.length,
        '>=',
        pageSize.value,
        '=>',
        hasMore.value,
      );
    if (newData.length > 0) {
      data.value.push(...newData);
    }

    // last page and each has more
    if (newData.length < pageSize.value && each.value.length > 0) {
      // use next each element
      each.value.shift();
      page.value = 1;
      pageSize.value = 0;
      hasMore.value = each.value.length > 0;
      name === 'secrets' && console.log('next', each.value?.[0] as S);
      loading.value = false;
      await loadData();
    }
    pageSize.value = newData.length;
    loading.value = false;
  }

  onMounted(loadData);
  watch(page, loadData);

  useInfiniteScroll(
    scrollElement,
    () => {
      if (isActive() && !loading.value && hasMore.value) {
        // load more
        page.value += 1;
      }
    },
    { distance: 10 },
  );

  async function resetPage() {
    hasMore.value = true;
    data.value = [];
    each.value = (_each ?? []) as UnwrapRef<S[]>;
    if (page.value !== 1) {
      // just set page = 1, will be handled by watcher
      page.value = 1;
    } else {
      // we need to reload, but page is already 1, so changing won't trigger watcher
      await loadData();
    }
  }

  function nextPage() {
    page.value += 1;
  }

  return { resetPage, nextPage, data, hasMore, loading };
}
