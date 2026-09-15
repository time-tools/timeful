<template>
  <v-scale-transition appear origin="center">
    <v-speed-dial
      v-model="speedDial"
      fixed
      transition="slide-y-reverse-transition"
      class="tw:bottom-4 tw:left-0 tw:right-0 tw:mx-auto tw:w-fit tw:sm:bottom-8"
    >
      <template #activator="{ props: activatorProps }">
        <v-btn
          v-bind="activatorProps"
          icon
          color="primary"
          class="tw:text-white"
        >
          <v-icon v-if="speedDial"><MdiClose /></v-icon>
          <v-icon v-else><MdiPlus /></v-icon>
        </v-btn>
      </template>
      <v-btn icon size="small" class="tw:bg-dark-gray tw:text-white">
        <v-icon small><MdiAccountGroup /></v-icon>
      </v-btn>
      <v-btn icon size="small" class="tw:bg-dark-gray tw:text-white">
        <v-icon small><MdiCalendarCursor /></v-icon>
      </v-btn>
    </v-speed-dial>
  </v-scale-transition>
</template>

<script setup lang="ts">
import { ref } from "vue"
import MdiAccountGroup from "~icons/mdi/account-group"
import MdiCalendarCursor from "~icons/mdi/calendar-cursor"
import MdiClose from "~icons/mdi/close"
import MdiPlus from "~icons/mdi/plus"

defineEmits<{
  createGroup: []
  createEvent: []
}>()

const speedDial = ref(false)
</script>
