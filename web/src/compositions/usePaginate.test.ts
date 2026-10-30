import { shallowMount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { nextTick, watch } from 'vue';
import type { Ref } from 'vue';

import { usePagination } from './usePaginate';

async function waitForState<T>(ref: Ref<T>, expected: T): Promise<void> {
  await new Promise<void>((resolve) => {
    watch(
      ref,
      (value) => {
        if (value === expected) {
          resolve();
        }
      },
      { immediate: true },
    );
  });
}

// eslint-disable-next-line promise/prefer-await-to-callbacks
export const mountComposition = (cb: () => void) => {
  const wrapper = shallowMount({
    setup() {
      // eslint-disable-next-line promise/prefer-await-to-callbacks
      cb();
      return {};
    },
    template: '<div />',
  });

  return wrapper;
};

describe('usePaginate', () => {
  const repoSecrets = [
    [{ name: 'repo1' }, { name: 'repo2' }, { name: 'repo3' }],
    [{ name: 'repo4' }, { name: 'repo5' }, { name: 'repo6' }],
  ];
  const orgSecrets = [
    [{ name: 'org1' }, { name: 'org2' }, { name: 'org3' }],
    [{ name: 'org4' }, { name: 'org5' }, { name: 'org6' }],
  ];

  it('should get first page', async () => {
    let usePaginationComposition = null as unknown as ReturnType<typeof usePagination>;
    mountComposition(() => {
      usePaginationComposition = usePagination<{ name: string }>(
        async (page) => repoSecrets[page - 1],
        () => true,
        { pageSize: 3 },
      );
    });
    await waitForState(usePaginationComposition.loading, true);
    await waitForState(usePaginationComposition.loading, false);

    expect(usePaginationComposition.data.value.length).toBe(3);
    expect(usePaginationComposition.data.value[0]).toStrictEqual(repoSecrets[0][0]);
  });

  it('should get first & second page', async () => {
    let usePaginationComposition = null as unknown as ReturnType<typeof usePagination>;
    mountComposition(() => {
      usePaginationComposition = usePagination<{ name: string }>(
        async (page) => repoSecrets[page - 1],
        () => true,
        { pageSize: 3 },
      );
    });
    await waitForState(usePaginationComposition.loading, true);
    await waitForState(usePaginationComposition.loading, false);

    usePaginationComposition.nextPage();
    await waitForState(usePaginationComposition.loading, false);

    expect(usePaginationComposition.data.value.length).toBe(6);
    expect(usePaginationComposition.data.value.at(-1)).toStrictEqual(repoSecrets[1][2]);
  });

  it('should get first page for each category', async () => {
    let usePaginationComposition = null as unknown as ReturnType<typeof usePagination>;
    mountComposition(() => {
      usePaginationComposition = usePagination<{ name: string }>(
        async (page, level) => {
          if (level === 'repo') {
            return repoSecrets[page - 1];
          }
          return orgSecrets[page - 1];
        },
        () => true,
        { each: ['repo', 'org'], pageSize: 3 },
      );
    });
    await waitForState(usePaginationComposition.loading, true);
    await waitForState(usePaginationComposition.loading, false);

    usePaginationComposition.nextPage();
    await waitForState(usePaginationComposition.loading, false);

    usePaginationComposition.nextPage();
    await waitForState(usePaginationComposition.loading, false);

    usePaginationComposition.nextPage();
    await waitForState(usePaginationComposition.loading, false);

    expect(usePaginationComposition.data.value.length).toBe(9);
    expect(usePaginationComposition.data.value.at(-1)).toStrictEqual(orgSecrets[0][2]);
  });

  it('should reset page and get first page again', async () => {
    let usePaginationComposition = null as unknown as ReturnType<typeof usePagination>;
    mountComposition(() => {
      usePaginationComposition = usePagination<{ name: string }>(
        async (page) => repoSecrets[page - 1],
        () => true,
        { pageSize: 3 },
      );
    });
    await waitForState(usePaginationComposition.loading, true);
    await waitForState(usePaginationComposition.loading, false);

    usePaginationComposition.nextPage();
    await waitForState(usePaginationComposition.loading, false);

    void usePaginationComposition.resetPage();
    await waitForState(usePaginationComposition.loading, false);

    expect(usePaginationComposition.data.value.length).toBe(3);
    expect(usePaginationComposition.data.value[0]).toStrictEqual(repoSecrets[0][0]);
  });

  it('should not hasMore when no data is left', async () => {
    let usePaginationComposition = null as unknown as ReturnType<typeof usePagination>;
    mountComposition(() => {
      usePaginationComposition = usePagination<{ name: string }>(
        async (page) => repoSecrets[page - 1],
        () => true,
        { pageSize: 3 },
      );
    });
    await waitForState(usePaginationComposition.loading, true);
    await waitForState(usePaginationComposition.loading, false);

    expect(usePaginationComposition.hasMore.value).toBe(true);
    expect(usePaginationComposition.data.value.length).toBe(3);

    usePaginationComposition.nextPage();
    await waitForState(usePaginationComposition.loading, false);
    expect(usePaginationComposition.hasMore.value).toBe(true);
    expect(usePaginationComposition.data.value.length).toBe(6);

    usePaginationComposition.nextPage();
    await waitForState(usePaginationComposition.loading, false);
    expect(usePaginationComposition.hasMore.value).toBe(false);
    expect(usePaginationComposition.data.value.length).toBe(6);
  });

  describe('reset while a request is in flight', () => {
    interface Deferred {
      page: number;
      resolve: (items: { name: string }[]) => void;
    }

    function useControlledPagination(pageSize = 3) {
      const calls: Deferred[] = [];
      let usePaginationComposition = null as unknown as ReturnType<typeof usePagination<{ name: string }>>;

      mountComposition(() => {
        usePaginationComposition = usePagination<{ name: string }>(
          async (page) =>
            new Promise<{ name: string }[]>((resolve) => {
              calls.push({ page, resolve });
            }),
          () => true,
          { pageSize },
        );
      });

      return { calls, composition: usePaginationComposition };
    }

    const flush = async () => {
      // let pending promise callbacks and (pre-flush) watchers run
      await nextTick();
      await nextTick();
    };

    it('discards a stale response that resolves after resetPage', async () => {
      // real world: user scrolls to page 2, then switches the branch filter
      // while the page 2 request is still in flight
      const { calls, composition } = useControlledPagination();

      await flush();
      calls[0].resolve([{ name: 'old1' }, { name: 'old2' }, { name: 'old3' }]);
      await waitForState(composition.loading, false);

      composition.nextPage();
      await flush();
      expect(calls[1].page).toBe(2);

      // switch filter: reset while page 2 request still pending
      void composition.resetPage();
      await flush();

      const freshCall = calls.at(-1)!;
      expect(freshCall.page).toBe(1);
      freshCall.resolve([{ name: 'new1' }, { name: 'new2' }, { name: 'new3' }]);
      await flush();

      // stale page 2 response arrives last
      calls[1].resolve([{ name: 'stale1' }, { name: 'stale2' }, { name: 'stale3' }]);
      await flush();

      expect(composition.data.value.map(({ name }) => name)).toStrictEqual(['new1', 'new2', 'new3']);
    });

    it('does not let a stale response overwrite hasMore', async () => {
      const { calls, composition } = useControlledPagination();

      await flush();
      calls[0].resolve([{ name: 'old1' }, { name: 'old2' }, { name: 'old3' }]);
      await waitForState(composition.loading, false);

      composition.nextPage();
      await flush();

      void composition.resetPage();
      await flush();

      const freshCall = calls.at(-1)!;
      freshCall.resolve([{ name: 'new1' }, { name: 'new2' }, { name: 'new3' }]);
      await flush();

      // stale response is a short (last) page and would clear hasMore
      calls[1].resolve([{ name: 'stale1' }]);
      await flush();

      expect(composition.hasMore.value).toBe(true);
      expect(composition.data.value.map(({ name }) => name)).toStrictEqual(['new1', 'new2', 'new3']);
    });

    it('reloads page 1 when resetPage is called while the initial request is in flight', async () => {
      // real world: user switches the filter twice quickly while still on page 1
      const { calls, composition } = useControlledPagination();

      await flush();
      expect(calls[0].page).toBe(1);

      void composition.resetPage();
      await flush();

      expect(calls.length).toBe(2);
      const freshCall = calls.at(-1)!;
      expect(freshCall.page).toBe(1);

      freshCall.resolve([{ name: 'new1' }, { name: 'new2' }, { name: 'new3' }]);
      await flush();

      // stale initial response arrives last
      calls[0].resolve([{ name: 'stale1' }, { name: 'stale2' }, { name: 'stale3' }]);
      await flush();

      expect(composition.data.value.map(({ name }) => name)).toStrictEqual(['new1', 'new2', 'new3']);
      expect(composition.loading.value).toBe(false);
    });

    it('keeps loading consistent so the next page can still be fetched after a stale response', async () => {
      const { calls, composition } = useControlledPagination();

      await flush();
      calls[0].resolve([{ name: 'old1' }, { name: 'old2' }, { name: 'old3' }]);
      await waitForState(composition.loading, false);

      composition.nextPage();
      await flush();

      void composition.resetPage();
      await flush();

      const freshCall = calls.at(-1)!;

      // stale page 2 response resolves while the fresh page 1 request is in flight;
      // it must not flip loading back to false mid-request
      calls[1].resolve([{ name: 'stale1' }, { name: 'stale2' }, { name: 'stale3' }]);
      await flush();
      expect(composition.loading.value).toBe(true);

      freshCall.resolve([{ name: 'new1' }, { name: 'new2' }, { name: 'new3' }]);
      await waitForState(composition.loading, false);

      composition.nextPage();
      await flush();
      const page2Call = calls.at(-1)!;
      expect(page2Call.page).toBe(2);
      page2Call.resolve([{ name: 'new4' }, { name: 'new5' }, { name: 'new6' }]);
      await waitForState(composition.loading, false);

      expect(composition.data.value.map(({ name }) => name)).toStrictEqual([
        'new1',
        'new2',
        'new3',
        'new4',
        'new5',
        'new6',
      ]);
    });
  });
});
