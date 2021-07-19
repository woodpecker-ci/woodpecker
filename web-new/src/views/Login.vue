<template>
  <div class="flex m-auto">
    <Button @click="doLogin">Login</Button>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, PropType } from 'vue';
import Button from '~/components/atomic/Button.vue';
import { authenticate, isAuthenticated } from '~/compositions/useAuthentication';
import router from '~/router';

export default defineComponent({
  name: 'Login',

  components: {
    Button,
  },

  props: {
    origin: {
      type: String as PropType<string | undefined>,
      default: null,
    },
  },

  setup(props) {
    function doLogin() {
      authenticate(props.origin);
    }

    onMounted(async () => {
      if (isAuthenticated()) {
        await router.replace({ name: 'home' });
      }
    });

    return {
      doLogin,
    };
  },
});
</script>
