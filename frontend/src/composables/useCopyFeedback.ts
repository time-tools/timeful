import {
  computed,
  getCurrentScope,
  onScopeDispose,
  ref,
  type ComputedRef,
  type Ref,
} from "vue"

export const COPY_FEEDBACK_DURATION_MS = 2000

export interface UseCopyFeedbackOptions {
  announcement?: string
}

export interface CopyFeedback {
  announcement: ComputedRef<string>
  copied: Ref<boolean>
  copy: (text: string) => Promise<void>
  reset: () => void
}

export function useCopyFeedback(
  options: UseCopyFeedbackOptions = {},
): CopyFeedback {
  const copied = ref(false)
  let timer: ReturnType<typeof setTimeout> | undefined

  function clearTimer() {
    if (timer === undefined) return
    clearTimeout(timer)
    timer = undefined
  }

  function reset() {
    clearTimer()
    copied.value = false
  }

  if (getCurrentScope()) onScopeDispose(reset)

  async function copy(text: string) {
    try {
      await navigator.clipboard.writeText(text)
    } catch (cause) {
      reset()
      throw cause
    }
    clearTimer()
    copied.value = true
    timer = setTimeout(() => {
      timer = undefined
      copied.value = false
    }, COPY_FEEDBACK_DURATION_MS)
  }

  const announcement = computed(() =>
    copied.value ? (options.announcement ?? "Copied") : "",
  )

  return { announcement, copied, copy, reset }
}
