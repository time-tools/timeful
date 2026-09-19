<template>
  <v-btn
    v-if="showUnarchive"
    variant="outlined"
    :disabled="busy"
    @click="toggleArchive"
  >
    <v-icon><MdiArchiveArrowUpOutline /></v-icon>
    <span class="tw:ml-1">Unarchive event</span>
  </v-btn>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import MdiArchiveArrowUpOutline from "~icons/mdi/archive-arrow-up-outline"
import type { Event } from "@/types"
import { canManageEventAsCurrentViewer } from "@/composables/event/eventOwnership"
import { archiveEvent } from "@/utils/services/EventService"
import { useMainStore } from "@/stores/main"

const props = defineProps<{ event: Event }>()
const emit = defineEmits<{ changed: [] }>()
const store = useMainStore()
const busy = ref(false)

const showUnarchive = computed(
  () =>
    props.event.isArchived === true &&
    canManageEventAsCurrentViewer(props.event),
)

async function toggleArchive() {
  if (!props.event._id || !showUnarchive.value || busy.value) return
  busy.value = true
  try {
    await archiveEvent(props.event._id, !props.event.isArchived)
    emit("changed")
  } catch {
    store.showError(
      "Could not update the event. Refresh the page and try again.",
    )
  } finally {
    busy.value = false
  }
}
</script>
