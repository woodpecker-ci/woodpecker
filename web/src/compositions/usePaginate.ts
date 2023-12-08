import { useInfiniteScroll } from '@vueuse/core';
import { onMounted, Ref, ref, watch } from 'vue';

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
  _loadData: (page: number, args: S) => Promise<T[] | null>,
  isActive: () => boolean,
  { scrollElement: _scrollElement, each: _each }: { scrollElement?: Ref<HTMLElement | null>; each?: S[] } = {},
) {
  const scrollElement = _scrollElement ?? ref(document.getElementById('scroll-component'));
  const page = ref(1);
  const pageSize = ref(0);
  const hasMore = ref(true);
  const data = ref<T[]>([]) as Ref<T[]>;
  const loading = ref(false);
  const each = ref<S[]>(_each || []);

  async function loadData() {
    loading.value = true;
    const newData = await _loadData(page.value, each.value?.[0] as S);
    hasMore.value = (newData !== null && newData.length >= pageSize.value) || each.value.length > 0;
    if (newData && newData.length > 0) {
      data.value.push(...newData);
      pageSize.value = newData.length;
    } else if (each.value.length > 0) {
      // use next each element
      each.value.shift();
      page.value = 1;
      pageSize.value = 0;
      hasMore.value = each.value.length > 0;
    }
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

  const resetPage = () => {
    if (page.value !== 1) {
      // just set page = 1, will be handled by watcher
      page.value = 1;
    } else {
      // we need to reload, but page is already 1, so changing won't trigger watcher
      loadData();
    }
  };

  return { resetPage, data };
}
