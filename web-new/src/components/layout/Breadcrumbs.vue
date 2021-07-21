<template>
  <div class="flex">
    <template v-for="(path, index) in paths">
      <span v-if="typeof path === 'string'">{{ path }}</span>
      <router-link v-else :to="path.link" class="text-link">{{ path.name }}</router-link>
      <span v-if="index !== Object.keys(paths).length - 1" class="mx-2"> / </span>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, PropType } from 'vue';
import { RouteLocationRaw } from 'vue-router';

type Breadcrumb =
  | string
  | {
      name: string;
      link: RouteLocationRaw;
    };

export default defineComponent({
  name: 'Breadcrumbs',

  props: {
    paths: {
      type: Array as PropType<Breadcrumb[]>,
      required: true,
    },
  },
});
</script>
