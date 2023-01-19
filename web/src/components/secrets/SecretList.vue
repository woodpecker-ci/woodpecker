<template>
  <div class="space-y-4 text-color">
    <ListItem v-for="secret in secrets" :key="secret.id" class="items-center">
      <span>{{ secret.name }}</span>

      <div v-if="secret.value" class="ml-auto">
        <span v-if="secret.showSecret" class="secret-value secret-value--visible" v-text="secret.value">
        </span>
        <span v-else class="secret-value secret-value--hidden">
          ****************
        </span>
        <IconButton
          icon="copy"
          class="ml-2 w-8 h-8 secret-action secret-action--copy"
          :class="{invisible: !secret.showSecret}"
          :title="$t('repo.settings.secrets.copy')"
          @click="copySecret(secret)"
        />
        <IconButton
          icon="show"
          class="ml-2 w-8 h-8 secret-action"
          :title="$t('repo.settings.secrets.toggle')"
          @click="toggleSecret(secret)"
        />
      </div>

      <div class="ml-auto">
        <span
          v-for="event in secret.event"
          :key="event"
          class="bg-gray-500 dark:bg-dark-700 dark:text-gray-400 text-white rounded-md mx-1 py-1 px-2 text-sm"
        >
          {{ event }}
        </span>
      </div>
      <IconButton
        icon="edit"
        class="ml-2 w-8 h-8"
        :title="$t('repo.settings.secrets.edit')"
        @click="editSecret(secret)"
      />
      <IconButton
        icon="trash"
        class="ml-2 w-8 h-8 hover:text-red-400 hover:dark:text-red-500"
        :is-loading="isDeleting"
        :title="$t('repo.settings.secrets.delete')"
        @click="deleteSecret(secret)"
      />
    </ListItem>

    <div v-if="secrets?.length === 0" class="ml-2">{{ $t(i18nPrefix + 'none') }}</div>
  </div>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { Secret } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Secret[];
  isDeleting: boolean;
  i18nPrefix: string;
}>();

const emit = defineEmits<{
  (event: 'edit', secret: Secret): void;
  (event: 'delete', secret: Secret): void;
}>();

const secrets = toRef(props, 'modelValue');

function editSecret(secret: Secret) {
  emit('edit', secret);
}

function deleteSecret(secret: Secret) {
  emit('delete', secret);
}

function toggleSecret(secret: Secret) {
  secret.showSecret = !secret.showSecret
}

function copySecret(secret: Secret) {
  navigator.clipboard.writeText(secret.value)
}
</script>

<style scoped>
.secret-value {
  width: 120px;
  max-width: 120px;
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: monospace;
  white-space: nowrap;
}

.secret-action {
  display: inline-block !important;
}

.secret-action--copy:active {
  border: 1px solid var(--fbc-secondary-text);
  border-radius: 100%;
  padding: 3px;
}

.secret-value--visible {
  color: var(--fbc-secondary-text);
}

.secret-value--hidden {
  opacity: 0.5;
}

.invisible {
  visibility: hidden;
}
</style>
