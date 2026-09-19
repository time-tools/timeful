<template>
  <template v-if="event.eventVisitorId && event._id">
    <v-btn variant="outlined" color="primary" @click="openDialog">
      <span class="tw:text-green">Manage access</span>
    </v-btn>
    <v-dialog
      v-model="dialog"
      max-width="540"
      scrollable
      :content-props="{ 'aria-labelledby': 'manage-access-title' }"
    >
      <v-card class="tw:pt-4">
        <EditorDialogHeader
          title="Manage access"
          title-id="manage-access-title"
          subtitle=""
          help-header=""
          :dialog="true"
          :show-help="false"
          :hide-dialog-actions="false"
          @close="dialog = false"
        />
        <v-card-text class="tw:flex tw:flex-col tw:gap-4 tw:px-4 tw:sm:px-8">
          <p class="tw:text-(--timeful-muted-foreground)">
            Use this event on another browser, or revoke access you granted
            earlier.
          </p>
          <v-alert v-if="error" type="error">{{ error }}</v-alert>
          <ol class="tw:flex tw:flex-col tw:gap-4">
            <li
              data-testid="manage-access-step-1"
              class="tw:flex tw:flex-col tw:gap-3 tw:rounded-lg tw:border tw:p-3"
              :class="stepClasses(1)"
              :aria-current="activeStep === 1 ? 'step' : undefined"
            >
              <h3 class="tw:flex tw:items-center tw:gap-2 tw:text-base">
                <span
                  class="tw:flex tw:h-6 tw:w-6 tw:shrink-0 tw:items-center tw:justify-center tw:rounded-full tw:text-xs tw:font-semibold"
                  :class="stepNumberClasses(1)"
                  aria-hidden="true"
                  >1</span
                >
                <span
                  ><span class="tw:sr-only">Step 1: </span>Create and copy a
                  transfer link</span
                >
              </h3>
              <p class="tw:text-sm tw:text-(--timeful-muted-foreground)">
                The link expires five minutes after you create it.
              </p>
              <div class="tw:flex tw:flex-wrap tw:justify-center tw:gap-2">
                <v-btn
                  :class="canCopy && 'tw:flex-1'"
                  :disabled="busyAction === 'start'"
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
              <v-text-field
                v-if="currentId"
                label="Transfer link"
                :model-value="link"
                readonly
              />
              <p
                role="status"
                class="tw:text-sm tw:text-(--timeful-muted-foreground)"
              >
                {{ currentId ? statusLabel : "" }}
              </p>
              <v-btn
                v-if="canCancel"
                :disabled="busyAction === 'cancel'"
                @click="cancel"
                >Cancel transfer</v-btn
              >
            </li>
            <li
              data-testid="manage-access-step-2"
              class="tw:flex tw:flex-col tw:gap-3 tw:rounded-lg tw:border tw:p-3"
              :class="stepClasses(2)"
              :aria-current="activeStep === 2 ? 'step' : undefined"
            >
              <h3 class="tw:flex tw:items-center tw:gap-2 tw:text-base">
                <span
                  class="tw:flex tw:h-6 tw:w-6 tw:shrink-0 tw:items-center tw:justify-center tw:rounded-full tw:text-xs tw:font-semibold"
                  :class="stepNumberClasses(2)"
                  aria-hidden="true"
                  >2</span
                >
                <span
                  ><span class="tw:sr-only">Step 2: </span>Open the link in the
                  other browser</span
                >
              </h3>
              <p class="tw:text-sm tw:text-(--timeful-muted-foreground)">
                It shows a matching code. Opening the link alone gives no
                access.
              </p>
            </li>
            <li
              data-testid="manage-access-step-3"
              class="tw:flex tw:flex-col tw:gap-3 tw:rounded-lg tw:border tw:p-3"
              :class="stepClasses(3)"
              :aria-current="activeStep === 3 ? 'step' : undefined"
            >
              <h3 class="tw:flex tw:items-center tw:gap-2 tw:text-base">
                <span
                  class="tw:flex tw:h-6 tw:w-6 tw:shrink-0 tw:items-center tw:justify-center tw:rounded-full tw:text-xs tw:font-semibold"
                  :class="stepNumberClasses(3)"
                  aria-hidden="true"
                  >3</span
                >
                <span
                  ><span class="tw:sr-only">Step 3: </span>Enter its code and
                  approve</span
                >
              </h3>
              <p
                v-if="store.authUser"
                class="tw:text-sm tw:text-(--timeful-muted-foreground)"
              >
                Approving signs the other browser in to your account.
              </p>
              <p v-else class="tw:text-sm tw:text-(--timeful-muted-foreground)">
                Approving grants access to your responses and, if you own this
                event, its owner controls.
              </p>
              <template v-if="isPending">
                <v-text-field
                  v-model="code"
                  label="Matching code from other browser"
                  autocomplete="off"
                  autocapitalize="characters"
                  spellcheck="false"
                  persistent-hint
                  hint="Type the code exactly as the other browser shows it."
                />
                <v-btn
                  :disabled="busyAction === 'approve' || !code"
                  @click="approve"
                  >Approve matching code</v-btn
                >
              </template>
              <p
                v-else-if="isApproved"
                class="tw:text-sm tw:text-(--timeful-muted-foreground)"
              >
                Approved — finish in the other browser.
              </p>
            </li>
          </ol>
          <section
            v-if="history.length"
            class="tw:flex tw:flex-col tw:gap-2"
            aria-labelledby="granted-access-title"
          >
            <h3 id="granted-access-title" class="tw:text-base">
              Granted access
            </h3>
            <div
              v-for="entry in history"
              :key="entry.id"
              class="tw:flex tw:items-center tw:gap-2"
            >
              <span>Granted access {{ entry.number }}</span>
              <v-btn
                :disabled="busyAction === 'revoke'"
                @click="revoke(entry.id)"
                >Revoke access</v-btn
              >
            </div>
          </section>
          <p aria-live="polite" class="tw:sr-only">
            {{ copied ? "Transfer link copied" : "" }}
          </p>
        </v-card-text>
      </v-card>
    </v-dialog>
  </template>
</template>
<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue"
import type { Event } from "@/types"
import { useMainStore } from "@/stores/main"
import EditorDialogHeader from "@/components/EditorDialogHeader.vue"
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
type PendingAction = "refresh" | "start" | "approve" | "cancel" | "revoke"
const dialog = ref(false)
const busy = ref(false)
const busyAction = ref<PendingAction>()
const copying = ref(false)
const polling = ref(false)
const error = ref("")
const copied = ref(false)
const currentId = ref("")
const current = ref<AccessTransfer>()
const code = ref("")
const history = ref<SavedTransfer[]>([])
const tracked = ref<SavedTransfer[]>([])
const statusLabels: Record<string, string> = {
  pending: "Waiting for the other browser to open the link.",
  approved: "Approved — finish in the other browser.",
  redeemed: "Completed.",
  cancelled: "Cancelled.",
  expired: "Expired — create a new link.",
  revoked: "Access revoked.",
  unavailable: "Unavailable — create a new link.",
}
const isPending = computed(() => current.value?.state === "pending")
const isApproved = computed(() => current.value?.state === "approved")
const canCancel = computed(() => isPending.value || isApproved.value)
const statusLabel = computed(() => {
  const transfer = current.value
  if (transfer?.state === "pending" && transfer.requests.length > 0)
    return "The other browser is showing a code — enter it in step 3."
  return (
    statusLabels[transfer?.state ?? "pending"] ??
    "Unavailable — create a new link."
  )
})
const activeStep = computed(() => {
  if (isPending.value) {
    if ((current.value?.requests.length ?? 0) > 0) return 3
    return copied.value ? 2 : 1
  }
  return isApproved.value ? 3 : 1
})
const link = computed(
  () =>
    `${window.location.origin}/transfer/${props.event._id}/${currentId.value}`,
)
const canCopy = computed(
  () =>
    current.value?.state === "pending" || current.value?.state === "approved",
)
let timer: ReturnType<typeof setInterval> | undefined

function stepClasses(index: number) {
  return index === activeStep.value
    ? "tw:border-(--timeful-outline-neutral) tw:bg-(--timeful-selection-bg)"
    : "tw:border-transparent"
}
function stepNumberClasses(index: number) {
  if (index === activeStep.value)
    return "tw:bg-(--timeful-primary-action-bg) tw:text-(--timeful-primary-action-fg)"
  if (index < activeStep.value)
    return "tw:bg-(--timeful-selection-bg) tw:text-(--timeful-selection-fg)"
  return "tw:border tw:border-(--timeful-outline-neutral) tw:text-(--timeful-muted-foreground)"
}
async function run(
  action: PendingAction,
  work: () => Promise<void>,
  failureMessage: string,
  clearError = true,
) {
  if (busy.value) return
  busy.value = true
  busyAction.value = action
  if (clearError) error.value = ""
  try {
    await work()
  } catch {
    if (clearError || !error.value) error.value = failureMessage
  } finally {
    busy.value = false
    busyAction.value = undefined
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
  void run(
    "refresh",
    refresh,
    "Could not refresh transfer status. Please try again.",
  )
}
async function start() {
  await run(
    "start",
    async () => {
      const eventId = props.event._id
      if (!eventId) return
      let retired: { entry: SavedTransfer; state: AccessTransfer } | undefined
      if (canCopy.value && currentId.value) {
        const transferId = currentId.value
        const entry = tracked.value.find(({ id }) => id === transferId)
        try {
          const state = await transferAction(eventId, transferId, "cancel")
          if (entry) retired = { entry, state }
        } catch (cause) {
          if (!isTransferUnavailable(cause)) throw cause
          // The target may have redeemed the transfer since the last poll.
          // Refresh before deciding whether the old transfer is still live.
          await refresh()
          if (canCopy.value) throw cause
        }
      }
      try {
        const created = await createTransfer(eventId)
        if (retired) updateTransfer(retired.entry, retired.state)
        current.value = created
        currentId.value = created.id
        tracked.value.push(rememberTransfer(eventId, currentId.value))
        code.value = ""
        copied.value = false
      } catch (cause) {
        if (retired) updateTransfer(retired.entry, retired.state)
        throw cause
      }
    },
    "Could not create a transfer link. Please try again.",
  )
}
async function copy() {
  if (busy.value || copying.value) return
  copying.value = true
  error.value = ""
  try {
    await navigator.clipboard.writeText(link.value)
    copied.value = true
  } catch {
    error.value =
      "Could not copy the transfer link. Select the link and copy it manually."
  } finally {
    copying.value = false
  }
}
async function approve() {
  await run(
    "approve",
    async () => {
      if (!props.event._id) return
      await refresh()
      const request =
        current.value && matchingRequest(current.value, code.value)
      if (!request) throw new Error("No matching target")
      current.value = await transferAction(
        props.event._id,
        currentId.value,
        "approve",
        { requestId: request.id, code: request.code },
      )
    },
    "Could not approve the transfer. Check the code or create a new link.",
  )
}
async function cancel() {
  await run(
    "cancel",
    async () => {
      if (props.event._id) {
        const entry = tracked.value.find(({ id }) => id === currentId.value)
        if (entry)
          updateTransfer(
            entry,
            await transferAction(props.event._id, entry.id, "cancel"),
          )
      }
    },
    "Could not cancel the transfer. Please try again.",
  )
}
async function revoke(id: string) {
  await run(
    "revoke",
    async () => {
      if (props.event._id) {
        const entry = tracked.value.find((entry) => entry.id === id)
        if (entry)
          updateTransfer(
            entry,
            await transferAction(props.event._id, id, "revoke"),
          )
      }
    },
    "Could not revoke granted access. Please try again.",
  )
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
