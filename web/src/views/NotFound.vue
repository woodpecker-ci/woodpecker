<template>
  <div class="flex h-full w-full flex-col items-center justify-center">
    <p class="mb-8 text-2xl">{{ $t('not_found.not_found') }}</p>
    <router-link class="text-wp-link-100 hover:text-wp-link-200" replace :to="{ name: 'home' }">
      {{ $t('not_found.back_home') }}
    </router-link>
  </div>
</template>
