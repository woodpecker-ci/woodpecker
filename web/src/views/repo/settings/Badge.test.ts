import { flushPromises, shallowMount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, nextTick, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { Repo, Token } from '~/lib/api/types';
import { RepoVisibility, TokenType } from '~/lib/api/types';

import Badge from './Badge.vue';

const apiClient = vi.hoisted(() => ({
  getRepoBranches: vi.fn(),
  getRepoToken: vi.fn(),
}));

vi.mock('~/compositions/useApiClient', () => ({ default: () => apiClient }));
vi.mock('~/compositions/useConfig', () => ({ default: () => ({ rootPath: '' }) }));
vi.mock('~/compositions/usePaginate', () => ({ usePaginate: async () => [] }));
vi.mock('~/compositions/useWPTitle', () => ({ useWPTitle: () => undefined }));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en },
});

const SettingsStub = defineComponent({ template: '<section><slot name="titleActions"/><slot/></section>' });
const InputFieldStub = defineComponent({ template: '<div><slot id="field"/></div>' });

function privateRepo(id: number): Repo {
  return {
    id,
    default_branch: 'main',
    full_name: `owner/repo-${id}`,
    visibility: RepoVisibility.Private,
  } as Repo;
}

function token(repoId: number, value: string): Token {
  return {
    id: repoId,
    repo_id: repoId,
    type: TokenType.Badge,
    value,
    created: 0,
  };
}

function mountBadge(repo = ref(privateRepo(1))) {
  const wrapper = shallowMount(Badge, {
    global: {
      plugins: [i18n],
      provide: { repo },
      stubs: {
        CheckboxesField: true,
        InputField: InputFieldStub,
        SelectField: true,
        Settings: SettingsStub,
        TextField: true,
      },
    },
  });
  return { repo, wrapper };
}

beforeEach(() => {
  apiClient.getRepoBranches.mockReset();
  apiClient.getRepoToken.mockReset();
  localStorage.clear();
});

describe('badge settings token', () => {
  it('hides private badge snippets until their token is loaded', async () => {
    const pending = Promise.withResolvers<Token>();
    apiClient.getRepoToken.mockReturnValueOnce(pending.promise);
    const { wrapper } = mountBadge();
    await nextTick();

    expect(wrapper.find('.code-box').exists()).toBe(false);

    pending.resolve(token(1, 'private-token'));
    await flushPromises();

    expect(wrapper.find('.code-box').text()).toContain('token=private-token');
  });

  it('discards a token response for a repo that is no longer displayed', async () => {
    const first = Promise.withResolvers<Token>();
    const second = Promise.withResolvers<Token>();
    apiClient.getRepoToken.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const { repo, wrapper } = mountBadge();
    await nextTick();

    repo.value = privateRepo(2);
    await nextTick();

    second.resolve(token(2, 'current-token'));
    await flushPromises();
    expect(wrapper.find('.code-box').text()).toContain('token=current-token');

    first.resolve(token(1, 'stale-token'));
    await flushPromises();
    expect(wrapper.find('.code-box').text()).toContain('token=current-token');
    expect(wrapper.find('.code-box').text()).not.toContain('token=stale-token');
  });
});
