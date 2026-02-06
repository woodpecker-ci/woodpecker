import type { InjectionKey, Ref } from 'vue';
import { inject as vueInject, provide as vueProvide } from 'vue';

import type { Org, OrgPermissions, Pipeline, PipelineConfig, Repo, RepoPermissions } from '~/lib/api/types';

import type { Tab } from './useTabs';

export interface InjectKeys {
  repo: Ref<Repo>;
  'repo-permissions': Ref<RepoPermissions>;
  org: Ref<Org>;
  'org-permissions': Ref<OrgPermissions>;
  pipeline: Ref<Pipeline>;
  'pipeline-configs': Ref<PipelineConfig[] | undefined>;
  tabs: Ref<Tab[]>;
  pipelines: Ref<Pipeline[]>;
}

export function requiredInject<T extends keyof InjectKeys>(key: T): InjectKeys[T] {
  const value = vueInject<InjectKeys[T]>(key);
  if (value === undefined) {
    throw new Error(`Unexpected: ${key} should be provided at this place`);
  }
  return value;
}

export function provide<T extends keyof InjectKeys>(key: T, value: InjectKeys[T]): void {
  return vueProvide(key, value as T extends InjectionKey<infer V> ? V : InjectKeys[T]);
}
