<template>
  <div
    class="tw:flex tw:items-baseline tw:gap-1 tw:rounded tw:bg-light-gray tw:p-2 tw:text-dark-gray"
  >
    <v-icon class="tw:text-base"><MdiAlertCircle /></v-icon>
    <span>
      <slot></slot>
    </span>
  </div>
</template>

<script setup lang="ts">
import MdiAlertCircle from "~icons/mdi/alert-circle"
</script>
