<template>
  <div class="w-full border border-gray-200 py-1 px-2 rounded-md bg-white hover:border-gray-300">
    <input
      v-bind="$attrs"
      v-model="innerValue"
      class="w-full text-gray-900 placeholder-gray-300 focus:outline-none focus:border-blue-400"
      type="text"
      :placeholder="placeholder"
    />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from 'vue';

export default defineComponent({
  name: 'TextField',

  components: {},

  props: {
    modelValue: {
      type: String,
      default: '',
    },

    placeholder: {
      type: String,
      default: '',
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: string): boolean => true,
  },

  setup: (props, ctx) => {
    const innerValue = computed({
      get: () => props.modelValue,
      set: (value) => {
        ctx.emit('update:modelValue', value);
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
