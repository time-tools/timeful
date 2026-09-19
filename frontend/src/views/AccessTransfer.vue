<template>
  <v-container class="tw:max-w-xl">
    <v-card>
      <v-card-title>
        <h1 class="tw:text-2xl tw:font-medium">Continue on this device</h1>
      </v-card-title>
      <v-card-text class="tw:flex tw:flex-col tw:gap-4">
        <v-alert v-if="error" type="error">{{ error }}</v-alert>
        <ol class="tw:flex tw:flex-col tw:gap-4">
          <li
            data-testid="access-transfer-step-2"
            class="tw:flex tw:flex-col tw:gap-3 tw:rounded-lg tw:border tw:p-3"
            :class="stepClasses(2)"
            :aria-current="!approved ? 'step' : undefined"
          >
            <h2 class="tw:flex tw:items-center tw:gap-2 tw:text-base">
              <span
                class="tw:flex tw:h-6 tw:w-6 tw:shrink-0 tw:items-center tw:justify-center tw:rounded-full tw:text-xs tw:font-semibold"
                :class="stepNumberClasses(2)"
                aria-hidden="true"
                >2</span
              >
              <span
                ><span class="tw:sr-only">Step 2: </span>Show this code to the
                browser that created the link</span
              >
            </h2>
            <p class="tw:text-sm tw:text-(--timeful-muted-foreground)">
              In that browser, choose Manage access, enter this six-digit code,
              and approve it within five minutes of creating the link.
            </p>
            <div
              v-if="code && !unavailable"
              class="tw:flex tw:flex-wrap tw:items-center tw:gap-3"
            >
              <p
                class="tw:text-3xl tw:font-bold tw:tracking-[0.2em]"
                data-testid="matching-code"
              >
                <span
                  v-for="(part, index) in codeGroups"
                  :key="index"
                  :class="{ 'tw:ml-3': index > 0 }"
                  >{{ part }}</span
                >
              </p>
              <v-btn
                :prepend-icon="copied ? MdiCheck : MdiContentCopy"
                @click="copyCode"
                >{{ copied ? "Copied" : "Copy code" }}</v-btn
              >
            </div>
            <p
              v-else-if="!error"
              role="status"
              class="tw:text-sm tw:text-(--timeful-muted-foreground)"
            >
              Loading the matching code…
            </p>
          </li>
          <li
            data-testid="access-transfer-step-3"
            class="tw:flex tw:flex-col tw:gap-3 tw:rounded-lg tw:border tw:p-3"
            :class="stepClasses(3)"
            :aria-current="approved ? 'step' : undefined"
          >
            <h2 class="tw:flex tw:items-center tw:gap-2 tw:text-base">
              <span
                class="tw:flex tw:h-6 tw:w-6 tw:shrink-0 tw:items-center tw:justify-center tw:rounded-full tw:text-xs tw:font-semibold"
                :class="stepNumberClasses(3)"
                aria-hidden="true"
                >3</span
              >
              <span
                ><span class="tw:sr-only">Step 3: </span>Continue after
                approval</span
              >
            </h2>
            <p role="status">
              {{
                approved
                  ? "Approved — you can continue"
                  : unavailable
                    ? "This link is expired, cancelled, or unavailable. Ask the source browser for a new link."
                    : "Waiting for approval in the other browser. You have no transferred access until then."
              }}
            </p>
            <v-btn v-if="code && !unavailable" :loading="busy" @click="finish()"
              >Continue after approval</v-btn
            >
          </li>
        </ol>
        <p aria-live="polite" class="tw:sr-only">
          {{ copyAnnouncement }}
        </p>
      </v-card-text>
    </v-card>
    <v-dialog
      v-model="confirmSwitch"
      max-width="480"
      persistent
      :content-props="{ 'aria-labelledby': 'switch-accounts-title' }"
    >
      <v-card>
        <v-card-title>
          <h2 id="switch-accounts-title" class="tw:text-xl">
            Switch accounts on this device?
          </h2>
        </v-card-title>
        <v-card-text>
          <p v-if="store.authUser">
            You are currently signed in as {{ store.authUser.firstName }}
            {{ store.authUser.lastName }} ({{ store.authUser.email }}).
          </p>
          <p>
            Continuing signs this browser in as the account that created the
            transfer, replacing your current sign-in. This does not merge
            accounts or make your current account another owner. Your current
            account's data and sign-ins on other devices stay intact, and you
            can sign back in.
          </p>
          <v-alert v-if="error" type="error">{{ error }}</v-alert>
        </v-card-text>
        <v-card-actions>
          <v-btn :disabled="busy" @click="confirmSwitch = false">Cancel</v-btn>
          <v-btn :loading="busy" @click="finish(true)">Switch accounts</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue"
import {
  isTransferUnavailable,
  requiresAccountSwitch,
  transferAction,
} from "@/composables/transfer/transferBoundary"
import { useMainStore } from "@/stores/main"
import { useCopyFeedback } from "@/composables/useCopyFeedback"
import MdiCheck from "~icons/mdi/check"
import MdiContentCopy from "~icons/mdi/content-copy"
const props = defineProps<{ eventId: string; transferId: string }>()
const store = useMainStore()
const code = ref("")
const approved = ref(false)
const unavailable = ref(false)
const UNAVAILABLE_MESSAGE =
  "This transfer is expired, cancelled, or unavailable. Ask the source browser for a new link."
const confirmSwitch = ref(false)
const error = ref("")
const busy = ref(false)
const {
  announcement: copyAnnouncement,
  copied,
  copy: copyToClipboard,
} = useCopyFeedback({ announcement: "Matching code copied" })
const polling = ref(false)
let timer: ReturnType<typeof setInterval> | undefined
const activeStep = computed(() => (approved.value ? 3 : 2))
const codeGroups = computed(() => [code.value.slice(0, 3), code.value.slice(3)])
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
function endWaitUnavailable() {
  unavailable.value = true
  error.value = UNAVAILABLE_MESSAGE
  if (timer) clearInterval(timer)
}
async function checkApproval() {
  if (polling.value || approved.value || unavailable.value || busy.value) return
  polling.value = true
  try {
    const transfer = await transferAction(
      props.eventId,
      props.transferId,
      "open",
    )
    if (transfer.state === "approved") {
      approved.value = true
      error.value = ""
    }
  } catch (cause) {
    // A definitive 403/404 means the link can no longer grant access; other
    // failures, such as a 500 or a network error, stay transient.
    if (isTransferUnavailable(cause)) endWaitUnavailable()
  } finally {
    polling.value = false
  }
}
watch([code, approved], ([currentCode, isApproved]) => {
  if (timer) clearInterval(timer)
  if (currentCode && !isApproved)
    timer = setInterval(() => {
      void checkApproval()
    }, 2000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
onMounted(async () => {
  try {
    const transfer = await transferAction(
      props.eventId,
      props.transferId,
      "open",
    )
    code.value = transfer.code
    approved.value = transfer.state === "approved"
  } catch {
    error.value = UNAVAILABLE_MESSAGE
  }
})
async function copyCode() {
  try {
    await copyToClipboard(code.value)
    error.value = ""
  } catch {
    error.value =
      "Could not copy the code. Select the code and copy it manually."
  }
}
async function finish(confirmAccountSwitch = false) {
  if (busy.value) return
  busy.value = true
  error.value = ""
  try {
    await transferAction(
      props.eventId,
      props.transferId,
      "redeem",
      confirmAccountSwitch ? { confirmAccountSwitch: true } : undefined,
    )
    window.location.assign(`/e/${props.eventId}`)
  } catch (cause) {
    if (requiresAccountSwitch(cause)) {
      confirmSwitch.value = true
      return
    }
    error.value =
      "Access has not been approved for this code, or the transfer is no longer available."
  } finally {
    busy.value = false
  }
}
</script>
