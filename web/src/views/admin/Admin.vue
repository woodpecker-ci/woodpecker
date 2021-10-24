<template>
  <div class="flex m-auto">Administration</div>
</template>

<script lang="ts">
import { defineComponent, onMounted } from 'vue';

import useAuthentication from '~/compositions/useAuthentication';
import router from '~/router';

export default defineComponent({
  name: 'Admin',

  setup() {
    const authentication = useAuthentication();

    onMounted(async () => {
      if (!authentication.user || !authentication.user.admin) {
        await router.replace({ name: 'home' });
      }
    });
  },
});
</script>
