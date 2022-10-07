import { inject as vueInject, provide as vueProvide, Ref } from 'vue';

import { Repo } from '~/lib/api/types';

export type InjectKeys = {
  repo: Ref<Repo>;
};

export function inject<T extends keyof InjectKeys>(key: T): InjectKeys[T] {
  const value = vueInject<InjectKeys[T]>(key);
  if (value === undefined) {
    throw new Error(`Please provide a value for ${key}`);
  }
  return value;
}

export function provide<T extends keyof InjectKeys>(key: T, value: InjectKeys[T]): void {
  return vueProvide(key, value);
}
