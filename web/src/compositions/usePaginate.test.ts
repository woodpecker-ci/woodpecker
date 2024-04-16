import { shallowMount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { type Ref, watch } from 'vue';

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
  it('get first repo page', async () => {
    const repoSecrets = [
      [{ name: 'repo1' }, { name: 'repo2' }, { name: 'repo3' }],
      [{ name: 'repo4' }, { name: 'repo5' }, { name: 'repo6' }],
    ];
    const orgSecrets = [
      [{ name: 'org1' }, { name: 'org2' }, { name: 'org3' }],
      [{ name: 'org4' }, { name: 'org5' }, { name: 'org6' }],
    ];

    let usePaginationComposition = null as unknown as ReturnType<typeof usePagination>;
    mountComposition(() => {
      usePaginationComposition = usePagination<{ name: string }>(
        async (page, level) => {
          console.log('getSingle', page, level);
          if (level === 'repo') {
            return repoSecrets[page - 1];
          }
          return orgSecrets[page - 1];
        },
        () => true,
        { each: ['repo', 'org'] },
      );
    });
    await waitForState(usePaginationComposition.loading, true);
    await waitForState(usePaginationComposition.loading, false);

    expect(usePaginationComposition.data.value.length).toBe(6);
  });
});
