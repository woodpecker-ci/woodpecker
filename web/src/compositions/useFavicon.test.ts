import { shallowMount } from '@vue/test-utils';
import { beforeAll, beforeEach, describe, expect, it } from 'vitest';
import { nextTick } from 'vue';

import type { PipelineStatus } from '~/lib/api/types';

import { useFavicon } from './useFavicon';

// eslint-disable-next-line promise/prefer-await-to-callbacks
const mountComposition = (cb: () => void) => {
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

function mountConsumer(status: PipelineStatus) {
  let favicon = null as unknown as ReturnType<typeof useFavicon>;
  const wrapper = mountComposition(() => {
    favicon = useFavicon();
    favicon.updateStatus(status);
  });
  return { wrapper, favicon };
}

function faviconHref(): string {
  return (document.getElementById('favicon-png') as HTMLLinkElement).href;
}

describe('useFavicon', () => {
  beforeAll(() => {
    const png = document.createElement('link');
    png.id = 'favicon-png';
    const svg = document.createElement('link');
    svg.id = 'favicon-svg';
    document.head.append(png, svg);
  });

  beforeEach(async () => {
    // start every test from a clean default state
    useFavicon().updateStatus('default');
    await nextTick();
  });

  it('reflects the pipeline status in the favicon', async () => {
    const { wrapper } = mountConsumer('running');
    await nextTick();

    expect(faviconHref()).toMatch(/favicon-(light|dark)-pending\.png$/);

    wrapper.unmount();
  });

  it('maps failure statuses to the error favicon', async () => {
    const { wrapper } = mountConsumer('failure');
    await nextTick();

    expect(faviconHref()).toMatch(/favicon-(light|dark)-error\.png$/);

    wrapper.unmount();
  });

  it('resets to the default favicon when the consuming component unmounts', async () => {
    // real world: user opens a failing pipeline, then navigates back to the
    // repo list — the error favicon must not stick around globally
    const { wrapper } = mountConsumer('failure');
    await nextTick();
    expect(faviconHref()).toMatch(/favicon-(light|dark)-error\.png$/);

    wrapper.unmount();
    await nextTick();

    expect(faviconHref()).toMatch(/favicon-(light|dark)-default\.png$/);
  });

  it('keeps the status of a still-mounted consumer when another one unmounts', async () => {
    // real world: overlapping consumers during a route transition
    const first = mountConsumer('running');
    const second = mountConsumer('success');
    await nextTick();
    expect(faviconHref()).toMatch(/favicon-(light|dark)-success\.png$/);

    first.wrapper.unmount();
    await nextTick();

    expect(faviconHref()).toMatch(/favicon-(light|dark)-success\.png$/);

    second.wrapper.unmount();
    await nextTick();
    expect(faviconHref()).toMatch(/favicon-(light|dark)-default\.png$/);
  });

  it('can be used outside a component scope without registering cleanup', async () => {
    // real world: imperative update from non-component code
    const favicon = useFavicon();
    favicon.updateStatus('running');
    await nextTick();

    expect(faviconHref()).toMatch(/favicon-(light|dark)-pending\.png$/);

    favicon.updateStatus('default');
    await nextTick();
    expect(faviconHref()).toMatch(/favicon-(light|dark)-default\.png$/);
  });
});
