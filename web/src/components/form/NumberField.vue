<template>
  <TextField v-model="innerValue" :placeholder="placeholder" type="number" />
</template>

<script lang="ts">
import { computed, defineComponent, toRef } from 'vue';

import TextField from '~/components/form/TextField.vue';

export default defineComponent({
  name: 'NumberField',

  components: { TextField },

  props: {
    modelValue: {
      type: Number,
      required: true,
    },

    placeholder: {
      type: String,
      default: '',
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: number): boolean => true,
  },

  setup: (props, ctx) => {
    const modelValue = toRef(props, 'modelValue');
    const innerValue = computed({
      get: () => modelValue.value.toString(),
      set: (value) => {
        ctx.emit('update:modelValue', parseFloat(value));
      },
    });

    return {
      innerValue,
    };
  },
});
</script>
