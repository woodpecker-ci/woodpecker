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
      default: '',
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

<style>
.form-control:focus {
  color: #495057;
  background-color: #fff;
  border-color: #1991eb;
  outline: 0;
  box-shadow: 0 0 0 2px rgb(70 127 207 / 25%);
}
.form-control {
  transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;
}
</style>
