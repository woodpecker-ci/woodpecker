import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { Org } from '~/lib/api/types';

import AdminOrgs from './AdminOrgs.vue';

const getOrgs = vi.fn<(query: { page: number }) => Promise<Org[] | null>>();
const remapOrg = vi.fn<(org: Org) => Promise<Org>>();

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({
    getOrgs,
    remapOrg,
    deleteOrg: vi.fn(),
  }),
}));

vi.mock('~/compositions/useNotifications', () => ({
  default: () => ({
    notify: vi.fn(),
    notifyError: vi.fn(),
  }),
}));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en },
});

describe('adminOrgs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('marks orgs with a name conflict and gives them to the new forge account', async () => {
    getOrgs.mockResolvedValue([
      { id: 1, name: 'acme', conflict_forge_remote_id: '42' } as Org,
      { id: 2, name: 'umbrella' } as Org,
    ]);
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    const wrapper = mount(AdminOrgs, { global: { plugins: [i18n], stubs: ['router-link'] } });
    await flushPromises();

    expect(wrapper.text()).toContain('Name conflict');
    const remapButtons = wrapper.findAll('button[title="Give to the new forge account"]');
    expect(remapButtons).toHaveLength(1);

    await remapButtons[0].trigger('click');
    await flushPromises();

    expect(remapOrg).toHaveBeenCalledWith(expect.objectContaining({ name: 'acme' }));
  });
});
