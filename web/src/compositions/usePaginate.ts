import { useInfiniteScroll } from '@vueuse/core';
import { onMounted, ref, watch } from 'vue';

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

export function usePagination<T>(
  _loadData: (page: number) => Promise<T[] | null>,
  isActive: () => boolean = () => true,
) {
  const page = ref(1);
  const pageSize = ref(0);
  const hasMore = ref(true);
  const data = ref<T[]>([]);
  const loading = ref(false);

  async function loadData() {
    loading.value = true;
    const newData = await _loadData(page.value);
    hasMore.value = newData != null && newData.length >= pageSize.value;
    if (newData != null) {
      if (page.value === 1) {
        pageSize.value = newData.length;
        data.value = newData;
      } else {
        data.value.push(...newData);
      }
    } else if (page.value === 1) {
      data.value = [];
    }
    loading.value = false;
  }

  onMounted(loadData);
  watch(page, loadData);

  useInfiniteScroll(
    document.getElementById('scroll-component'),
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
