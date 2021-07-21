<template>
  <div class="flex m-auto">
    <Button @click="doLogin">Login</Button>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, PropType } from 'vue';
import Button from '~/components/atomic/Button.vue';
import useAuthentication from '~/compositions/useAuthentication';
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
    const authentication = useAuthentication();

    function doLogin() {
      authentication.authenticate(props.origin);
    }

    onMounted(async () => {
      if (authentication.isAuthenticated) {
        await router.replace({ name: 'home' });
      }
    });

    return {
      doLogin,
    };
  },
});
</script>
