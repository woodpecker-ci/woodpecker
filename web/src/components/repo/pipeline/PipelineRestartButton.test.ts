import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { createI18n } from 'vue-i18n';

import PipelineRestartButton from '~/components/repo/pipeline/PipelineRestartButton.vue';

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en: {} },
});

function mountButton(canRestartFailed: boolean) {
  return mount(PipelineRestartButton, {
    props: { canRestartFailed },
    global: { plugins: [i18n] },
    attachTo: document.body,
  });
}

describe('pipelineRestartButton', () => {
  it('is a plain restart button when there is nothing failed to restart', async () => {
    const wrapper = mountButton(false);

    expect(wrapper.find('[data-test="pipeline-restart-options"]').exists()).toBe(false);
    await wrapper.find('[data-test="pipeline-restart"]').trigger('click');
    expect(wrapper.emitted('restart')).toHaveLength(1);
    wrapper.unmount();
  });

  it('offers restarting the failed workflows in its menu', async () => {
    const wrapper = mountButton(true);
    const toggle = wrapper.find('[data-test="pipeline-restart-options"]');

    expect(wrapper.find('[role="menu"]').exists()).toBe(false);
    expect(toggle.attributes('aria-expanded')).toBe('false');

    await toggle.trigger('click');
    expect(wrapper.find('[role="menu"]').exists()).toBe(true);
    expect(toggle.attributes('aria-expanded')).toBe('true');

    await wrapper.find('[data-test="pipeline-restart-failed"]').trigger('click');
    expect(wrapper.emitted('restartFailed')).toHaveLength(1);
    expect(wrapper.emitted('restart')).toBeUndefined();
    expect(wrapper.find('[role="menu"]').exists()).toBe(false);
    wrapper.unmount();
  });

  it('closes the menu on escape', async () => {
    const wrapper = mountButton(true);

    await wrapper.find('[data-test="pipeline-restart-options"]').trigger('click');
    await wrapper.find('[role="menu"]').trigger('keydown', { key: 'Escape' });
    expect(wrapper.find('[role="menu"]').exists()).toBe(false);
    wrapper.unmount();
  });
});
