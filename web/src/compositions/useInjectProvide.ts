import { inject as vueInject, InjectionKey, provide as vueProvide, Ref } from 'vue';

import { Org, OrgPermissions, Repo } from '~/lib/api/types';

export type InjectKeys = {
  repo: Ref<Repo>;
  org: Ref<Org | undefined>;
  'org-permissions': Ref<OrgPermissions | undefined>;
};

export function inject<T extends keyof InjectKeys>(key: T): InjectKeys[T] {
  const value = vueInject<InjectKeys[T]>(key);
  if (value === undefined) {
    throw new Error(`Please provide a value for ${key}`);
  }
  return value;
}

export function provide<T extends keyof InjectKeys>(key: T, value: InjectKeys[T]): void {
  return vueProvide(key, value as T extends InjectionKey<infer V> ? V : InjectKeys[T]);
}
