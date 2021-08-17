<template>
  <div class="flex w-full h-full justify-center items-center">
    <Panel class="flex flex-col m-8 md:flex-row md:w-3xl md:h-sm p-0 overflow-hidden">
      <div class="flex bg-green md:w-3/5 justify-center items-center">
        <img class="w-48 h-48" src="../assets/logo.svg" />
      </div>
      <div class="flex flex-col md:w-2/5 my-8 p-4 items-center justify-center">
        <h1 class="text-xl">Welcome to Woodpecker</h1>
        <Button class="mt-4" @click="doLogin">Login with SSO</Button>
      </div>
    </Panel>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, PropType } from 'vue';
import Button from '~/components/atomic/Button.vue';
import useAuthentication from '~/compositions/useAuthentication';
import { useRouter } from 'vue-router';
import Panel from '~/components/layout/Panel.vue';

export default defineComponent({
  name: 'Login',

  components: {
    Button,
    Panel,
  },

  props: {
    origin: {
      type: String as PropType<string | undefined>,
      default: null,
    },
  },

  setup(props) {
    const router = useRouter();
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
