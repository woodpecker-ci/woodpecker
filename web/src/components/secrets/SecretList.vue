<template>
  <div class="space-y-4 text-color">
    <ListItem v-for="secret in secrets" :key="secret.id" class="items-center">
      <span>{{ secret.name }}</span>
      <div class="ml-auto">
        <span
          v-for="event in secret.event"
          :key="event"
          class="bg-gray-500 dark:bg-dark-700 dark:text-gray-400 text-white rounded-md mx-1 py-1 px-2 text-sm"
        >
          {{ event }}
        </span>
      </div>
      <IconButton icon="edit" class="ml-2 w-8 h-8" @click="editSecret(secret)" />
      <IconButton
        icon="trash"
        class="ml-2 w-8 h-8 hover:text-red-400 hover:dark:text-red-500"
        :is-loading="isDeleting"
        @click="deleteSecret(secret)"
      />
    </ListItem>

    <div v-if="secrets?.length === 0" class="ml-2">{{ $t(i18nPrefix + 'none') }}</div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, toRef } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { Secret } from '~/lib/api/types';

export default defineComponent({
  name: 'SecretList',

  components: {
    ListItem,
    IconButton,
  },

  props: {
    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    modelValue: {
      type: Array as PropType<Secret[]>,
      required: true,
    },

    isDeleting: {
      type: Boolean,
      required: true,
    },

    i18nPrefix: {
      type: String,
      required: true,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    edit: (secret: Secret): boolean => true,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    delete: (secret: Secret): boolean => true,
  },

  setup(props, ctx) {
    const secrets = toRef(props, 'modelValue');

    function editSecret(secret: Secret) {
      ctx.emit('edit', secret);
    }

    function deleteSecret(secret: Secret) {
      ctx.emit('delete', secret);
    }

    return { secrets, editSecret, deleteSecret };
  },
});
</script>
