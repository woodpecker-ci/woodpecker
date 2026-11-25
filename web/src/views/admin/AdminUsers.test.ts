import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { nextTick } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { Forge, User } from '~/lib/api/types';

import AdminUsers from './AdminUsers.vue';

const forges: Forge[] = [
  { id: 1, type: 'github', url: 'https://github.com' },
  { id: 2, type: 'gitlab', url: 'https://gitlab.com' },
];

const getForges = vi.fn<(query: { page: number }) => Promise<Forge[]>>();
const getUsers = vi.fn<(query: { page: number }) => Promise<User[] | null>>();
const createUser = vi.fn<(user: Partial<User>) => Promise<User>>();

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({
    getForges,
    getUsers,
    createUser,
    updateUser: vi.fn(),
    deleteUser: vi.fn(),
  }),
}));

vi.mock('~/compositions/useNotifications', () => ({
  default: () => ({
    notify: vi.fn(),
    notifyError: vi.fn(),
  }),
}));

// resolve the labels through the real locale file, so renaming a key there
// fails this test instead of silently rendering the raw key in the app
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en },
});

const global = { plugins: [i18n] };

async function mountAdminUsers() {
  const wrapper = mount(AdminUsers, { global });
  await flushPromises();
  return wrapper;
}

async function openAddUserForm(wrapper: Awaited<ReturnType<typeof mountAdminUsers>>) {
  const addUserButton = wrapper.findAll('button').find((button) => button.text() === 'Add user');
  expect(addUserButton).toBeDefined();
  await addUserButton!.trigger('click');
  await nextTick();
}

describe('adminUsers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getForges.mockResolvedValue(forges);
    getUsers.mockResolvedValue([]);
    createUser.mockResolvedValue({} as User);
  });

  it('shows a forge select defaulting to the default forge when adding a user', async () => {
    const wrapper = await mountAdminUsers();
    await openAddUserForm(wrapper);

    const forgeSelect = wrapper.find('select');
    expect(forgeSelect.exists()).toBe(true);

    const options = forgeSelect.findAll('option');
    expect(options).toHaveLength(2);
    expect(forgeSelect.element.value).toBe('1');
    expect(options[1].text()).toBe('gitlab.com');
  });

  it('keeps the forge select after picking a forge for a new user', async () => {
    const wrapper = await mountAdminUsers();
    await openAddUserForm(wrapper);

    await wrapper.find('select').setValue('2');

    const forgeSelect = wrapper.find('select');
    expect(forgeSelect.exists()).toBe(true);
    expect(forgeSelect.element.value).toBe('2');
  });

  it('shows the forge of an existing user read-only', async () => {
    getUsers.mockResolvedValue([{ id: 7, login: 'octocat', forge_id: 2 } as User]);
    const wrapper = await mountAdminUsers();

    const editButton = wrapper.find('button[title="Edit user"]');
    expect(editButton.exists()).toBe(true);
    await editButton.trigger('click');
    await nextTick();

    expect(wrapper.find('select').exists()).toBe(false);
    const forgeInput = wrapper.findAll('input').find((input) => input.element.value === 'gitlab.com');
    expect(forgeInput).toBeDefined();
    expect(forgeInput!.element.disabled).toBe(true);
  });

  it('marks where the admin of a listed user comes from', async () => {
    getUsers.mockResolvedValue([
      { id: 1, login: 'db-admin', forge_id: 1, admin: true, admin_env: false } as User,
      { id: 2, login: 'env-admin', forge_id: 1, admin: true, admin_env: true } as User,
      { id: 3, login: 'revoked', forge_id: 1, admin: false, admin_env: true, admin_revoked: true } as User,
      { id: 4, login: 'user', forge_id: 1, admin: false, admin_env: false } as User,
      { id: 5, login: 'not-signed-in', forge_id: 1, admin: false, admin_env: true } as User,
    ]);
    const wrapper = await mountAdminUsers();

    const rowOf = (login: string) =>
      wrapper.findAll('span').find((span) => span.text() === login)!.element.parentElement!;
    // the forge badge has no hover text, the admin badges do
    const badgesOf = (login: string) =>
      [...rowOf(login).querySelectorAll('span[title]')].map((badge) => ({
        text: badge.textContent?.trim(),
        title: badge.getAttribute('title'),
      }));
    const adminTexts = (login: string) => rowOf(login).textContent;

    expect(adminTexts('db-admin')).toContain('Admin');
    expect(badgesOf('db-admin')).toEqual([]);

    expect(badgesOf('env-admin')).toEqual([
      { text: 'Admin (env)', title: 'Admin because the user is listed in WOODPECKER_ADMIN.' },
    ]);
    expect(badgesOf('revoked')).toEqual([
      {
        text: 'Admin temporarily revoked',
        title:
          'The user is listed in WOODPECKER_ADMIN environment configuration, but admin was revoked here. They are admin again at their first sign in after the next server restart, unless they are removed from WOODPECKER_ADMIN.',
      },
    ]);

    expect(adminTexts('user')).not.toContain('Admin');

    // a listed user gets admin at the first sign in, nothing was revoked
    expect(badgesOf('not-signed-in')).toEqual([
      { text: 'Admin (env)', title: 'Admin because the user is listed in WOODPECKER_ADMIN.' },
    ]);
  });

  it('creates the user on the forge picked in the form', async () => {
    const wrapper = await mountAdminUsers();
    await openAddUserForm(wrapper);

    await wrapper.find('select').setValue('2');

    const loginInput = wrapper.findAll('input').find((input) => input.element.placeholder === 'Username');
    expect(loginInput).toBeDefined();
    await loginInput!.setValue('octocat');

    await wrapper.find('form').trigger('submit');
    await flushPromises();

    expect(createUser).toHaveBeenCalledWith(expect.objectContaining({ login: 'octocat', forge_id: 2 }));
  });
});
