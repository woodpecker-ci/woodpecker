<template>
  <FluidContainer>
    <div class="flex border-b items-center pb-4 mb-4">
      <IconButton :to="{ name: 'repo' }" icon="back" />
      <h1 class="text-xl ml-2">Settings</h1>
    </div>

    <Tabs>
      <Tab title="General">
        <GeneralTab />
      </Tab>
      <Tab title="Secrets">
        <SecretsTab />
      </Tab>
      <Tab title="Registries">
        <RegistriesTab />
      </Tab>
      <Tab title="Badge">
        <BadgeTab />
      </Tab>
    </Tabs>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import Icon from '~/components/atomic/Icon.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import Tab from '~/components/tabs/Tab.vue';
import GeneralTab from '~/components/repo/settings/GeneralTab.vue';
import SecretsTab from '~/components/repo/settings/SecretsTab.vue';
import RegistriesTab from '~/components/repo/settings/RegistriesTab.vue';
import BadgeTab from '~/components/repo/settings/BadgeTab.vue';

export default defineComponent({
  name: 'RepoSettings',

  components: {
    FluidContainer,
    Button,
    IconButton,
    Icon,
    Tabs,
    Tab,
    GeneralTab,
    SecretsTab,
    RegistriesTab,
    BadgeTab,
  },
});
</script>
