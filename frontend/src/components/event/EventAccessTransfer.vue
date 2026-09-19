<template>
  <template v-if="event.eventVisitorId && event._id">
    <v-btn variant="outlined" color="primary" @click="openDialog">
      <span class="tw:text-green">Manage access</span>
    </v-btn>
    <v-dialog v-model="dialog" max-width="540">
      <v-card title="Manage access">
        <v-card-text class="tw:flex tw:flex-col tw:gap-4">
          <p>
            Use this event on another browser, or revoke access you granted
            earlier. Create a transfer link and copy it, open it on the other
            browser, and approve the matching code shown there within five
            minutes. Opening the link alone gives no access.
          </p>
          <p v-if="store.authUser">
            This signs the other browser in to your account.
          </p>
          <p v-else>
            This grants access to your responses and, if you own this event, its
            owner controls.
          </p>
          <v-alert v-if="error" type="error">{{ error }}</v-alert>
          <div class="tw:flex tw:flex-wrap tw:justify-center tw:gap-2">
            <v-btn
              :class="canCopy && 'tw:flex-1'"
              :disabled="busy"
              :prepend-icon="MdiRefresh"
              @click="start"
              >Create new transfer link</v-btn
            >
            <v-btn
              v-if="canCopy"
              class="tw:flex-1"
              :prepend-icon="MdiContentCopy"
              @click="copy"
              >{{ copied ? "Copied" : "Copy link" }}</v-btn
            >
          </div>
          <template v-if="currentId">
            <v-text-field label="Transfer link" :model-value="link" readonly />
            <p role="status">Transfer status: {{ statusLabel }}</p>
            <template v-if="current?.state === 'pending'">
              <v-text-field
                v-model="code"
                label="Matching code from other browser"
                autocomplete="off"
              />
              <v-btn :disabled="busy || !code" @click="approve"
                >Approve matching code</v-btn
              >
            </template>
            <v-btn
              v-if="
                current?.state === 'pending' || current?.state === 'approved'
              "
              :disabled="busy"
              @click="cancel"
              >Cancel transfer</v-btn
            >
          </template>
          <div
            v-for="entry in history"
            :key="entry.id"
            class="tw:flex tw:items-center tw:gap-2"
          >
            <span>Granted access {{ entry.number }}</span>
            <v-btn :disabled="busy" @click="revoke(entry.id)"
              >Revoke access</v-btn
            >
          </div>
        </v-card-text>
        <v-card-actions
          ><v-btn @click="dialog = false">Close</v-btn></v-card-actions
        >
      </v-card>
    </v-dialog>
  </template>
</template>
<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue"
import type { Event } from "@/types"
import { useMainStore } from "@/stores/main"
import {
  createTransfer,
  transferAction,
  matchingRequest,
  rememberTransfer,
  savedTransfers,
  forgetTransfer,
  isTransferUnavailable,
  type AccessTransfer,
  type SavedTransfer,
} from "@/composables/transfer/transferBoundary"
import MdiContentCopy from "~icons/mdi/content-copy"
import MdiRefresh from "~icons/mdi/refresh"

const props = defineProps<{ event: Event }>()
const store = useMainStore()
const dialog = ref(false)
const busy = ref(false)
const polling = ref(false)
const error = ref("")
const copied = ref(false)
const currentId = ref("")
const current = ref<AccessTransfer>()
const code = ref("")
const history = ref<SavedTransfer[]>([])
const tracked = ref<SavedTransfer[]>([])
const statusLabels: Record<string, string> = {
  pending: "Waiting for approval",
  approved: "Approved — waiting for the other browser",
  redeemed: "Completed",
  cancelled: "Cancelled",
  expired: "Expired — create a new link",
  revoked: "Access revoked",
  unavailable: "Unavailable — create a new link",
}
const statusLabel = computed(
  () =>
    statusLabels[current.value?.state ?? "pending"] ??
    "Unavailable — create a new link",
)
const link = computed(
  () =>
    `${window.location.origin}/transfer/${props.event._id}/${currentId.value}`,
)
const canCopy = computed(
  () =>
    current.value?.state === "pending" || current.value?.state === "approved",
)
let timer: ReturnType<typeof setInterval> | undefined

async function run(
  action: () => Promise<void>,
  failureMessage: string,
  clearError = true,
) {
  if (busy.value) return
  busy.value = true
  if (clearError) error.value = ""
  try {
    await action()
  } catch {
    if (clearError || !error.value) error.value = failureMessage
  } finally {
    busy.value = false
  }
}
function updateTransfer(entry: SavedTransfer, state: AccessTransfer) {
  if (!tracked.value.some(({ id }) => id === entry.id)) return
  if (entry.id === currentId.value) current.value = state
  history.value = history.value.filter(({ id }) => id !== entry.id)
  if (state.revocable) {
    history.value.push(entry)
    history.value.sort((a, b) => a.number - b.number)
  } else if (state.state !== "pending" && state.state !== "approved") {
    tracked.value = tracked.value.filter(({ id }) => id !== entry.id)
    if (props.event._id) forgetTransfer(props.event._id, entry.id)
  }
}
async function refresh() {
  const eventId = props.event._id
  if (!eventId) return
  let failed = false
  const states = new Map<string, AccessTransfer>()
  await Promise.all(
    tracked.value.map(async (entry) => {
      try {
        const state = await transferAction(eventId, entry.id, "status")
        states.set(entry.id, state)
        updateTransfer(entry, state)
      } catch (cause) {
        if (isTransferUnavailable(cause)) {
          const state: AccessTransfer = {
            id: entry.id,
            state: "unavailable",
            revocable: false,
            requestId: "",
            code: "",
            requests: [],
            confirmationRequired: false,
          }
          states.set(entry.id, state)
          updateTransfer(entry, state)
        } else {
          failed = true
        }
      }
    }),
  )
  if (!currentId.value) {
    const active = [...tracked.value]
      .sort((a, b) => b.number - a.number)
      .find((entry) => {
        const state = states.get(entry.id)?.state
        return state === "pending" || state === "approved"
      })
    const state = active ? states.get(active.id) : undefined
    if (active && state) {
      currentId.value = active.id
      current.value = state
    }
  }
  if (failed) throw new Error("Status refresh failed")
}
async function poll() {
  if (busy.value || polling.value) return
  polling.value = true
  try {
    await refresh()
  } catch {
    if (!error.value)
      error.value = "Could not refresh transfer status. Please try again."
  } finally {
    polling.value = false
  }
}
function openDialog() {
  if (props.event._id) {
    for (const entry of savedTransfers(props.event._id)) {
      if (!tracked.value.some(({ id }) => id === entry.id))
        tracked.value.push(entry)
    }
  }
  dialog.value = true
  void run(refresh, "Could not refresh transfer status. Please try again.")
}
async function start() {
  await run(async () => {
    const eventId = props.event._id
    if (!eventId) return
    if (canCopy.value && currentId.value) {
      const transferId = currentId.value
      const entry = tracked.value.find(({ id }) => id === transferId)
      try {
        const state = await transferAction(eventId, transferId, "cancel")
        if (entry) updateTransfer(entry, state)
      } catch (cause) {
        if (!isTransferUnavailable(cause)) throw cause
        // The target may have redeemed the transfer since the last poll.
        // Refresh before deciding whether the old transfer is still live.
        await refresh()
        if (canCopy.value) throw cause
      }
    }
    current.value = await createTransfer(eventId)
    currentId.value = current.value.id
    tracked.value.push(rememberTransfer(eventId, currentId.value))
    code.value = ""
    copied.value = false
  }, "Could not create a transfer link. Please try again.")
}
async function copy() {
  await run(async () => {
    await navigator.clipboard.writeText(link.value)
    copied.value = true
  }, "Could not copy the transfer link. Select the link and copy it manually.")
}
async function approve() {
  await run(async () => {
    if (!props.event._id) return
    await refresh()
    const request = current.value && matchingRequest(current.value, code.value)
    if (!request) throw new Error("No matching target")
    current.value = await transferAction(
      props.event._id,
      currentId.value,
      "approve",
      { requestId: request.id, code: request.code },
    )
  }, "Could not approve the transfer. Check the code or create a new link.")
}
async function cancel() {
  await run(async () => {
    if (props.event._id) {
      const entry = tracked.value.find(({ id }) => id === currentId.value)
      if (entry)
        updateTransfer(
          entry,
          await transferAction(props.event._id, entry.id, "cancel"),
        )
    }
  }, "Could not cancel the transfer. Please try again.")
}
async function revoke(id: string) {
  await run(async () => {
    if (props.event._id) {
      const entry = tracked.value.find((entry) => entry.id === id)
      if (entry)
        updateTransfer(
          entry,
          await transferAction(props.event._id, id, "revoke"),
        )
    }
  }, "Could not revoke granted access. Please try again.")
}
watch(dialog, (open) => {
  if (timer) clearInterval(timer)
  if (open)
    timer = setInterval(() => {
      void poll()
    }, 2000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
