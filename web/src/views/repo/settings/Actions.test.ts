import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { Repo, Token } from '~/lib/api/types';
import { RepoVisibility, TokenType } from '~/lib/api/types';

import Actions from './Actions.vue';

const apiClient = vi.hoisted(() => ({
  rotateRepoTokens: vi.fn(),
}));
const notifications = vi.hoisted(() => ({ notify: vi.fn() }));
const router = vi.hoisted(() => ({ replace: vi.fn() }));

vi.mock('vue-router', () => ({ useRouter: () => router }));
vi.mock('~/compositions/useApiClient', () => ({ default: () => apiClient }));
vi.mock('~/compositions/useNotifications', () => ({ default: () => notifications }));
vi.mock('~/compositions/useWPTitle', () => ({ useWPTitle: () => undefined }));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en },
});

const SettingsStub = defineComponent({ template: '<section><slot/></section>' });

function repo(): Repo {
  return {
    id: 1,
    active: true,
    forge_remote_id: '1',
    full_name: 'owner/repo',
    visibility: RepoVisibility.Private,
  } as Repo;
}

function token(value: string): Token {
  return {
    id: 1,
    repo_id: 1,
    type: TokenType.Badge,
    value,
    created: 0,
  };
}

function mountActions() {
  const badgeToken = ref('old-token');
  const wrapper = mount(Actions, {
    global: {
      plugins: [i18n],
      provide: { repo: ref(repo()), 'badge-token': badgeToken },
      stubs: { Settings: SettingsStub },
    },
  });
  const rotateButton = wrapper
    .findAll('button')
    .find((button) => button.text() === i18n.global.t('repo.settings.actions.rotate_tokens.rotate_tokens'));
  expect(rotateButton).toBeDefined();
  expect(rotateButton!.text()).toBe(i18n.global.t('repo.settings.actions.rotate_tokens.rotate_tokens'));
  return { badgeToken, rotateButton: rotateButton!, wrapper };
}

beforeEach(() => {
  apiClient.rotateRepoTokens.mockReset();
  notifications.notify.mockReset();
  router.replace.mockReset();
  vi.unstubAllGlobals();
});

describe('repository token rotation', () => {
  it('does not rotate tokens when confirmation is rejected', async () => {
    const confirm = vi.fn().mockReturnValue(false);
    vi.stubGlobal('confirm', confirm);
    const { badgeToken, rotateButton } = mountActions();

    await rotateButton.trigger('click');
    await flushPromises();

    expect(confirm).toHaveBeenCalledWith(i18n.global.t('repo.settings.actions.rotate_tokens.confirm'));
    expect(confirm.mock.calls[0][0]).toContain('stop working immediately');
    expect(apiClient.rotateRepoTokens).not.toHaveBeenCalled();
    expect(badgeToken.value).toBe('old-token');
  });

  it('updates the shared badge token from the rotation response', async () => {
    vi.stubGlobal('confirm', vi.fn().mockReturnValue(true));
    apiClient.rotateRepoTokens.mockResolvedValue([token('new-token')]);
    const { badgeToken, rotateButton } = mountActions();

    await rotateButton.trigger('click');
    await flushPromises();

    expect(apiClient.rotateRepoTokens).toHaveBeenCalledWith(1);
    expect(badgeToken.value).toBe('new-token');
    expect(notifications.notify).toHaveBeenCalledWith({
      title: i18n.global.t('repo.settings.actions.rotate_tokens.success'),
      type: 'success',
    });
  });
});
