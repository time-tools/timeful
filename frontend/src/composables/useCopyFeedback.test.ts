import { effectScope } from "vue"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { COPY_FEEDBACK_DURATION_MS, useCopyFeedback } from "./useCopyFeedback"

const writeText = vi.fn()

beforeEach(() => {
  vi.useFakeTimers()
  writeText.mockReset().mockResolvedValue(undefined)
  vi.stubGlobal("navigator", { clipboard: { writeText } })
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.useRealTimers()
})

describe("useCopyFeedback", () => {
  it("confirms success and automatically reverts after the feedback duration", async () => {
    const feedback = useCopyFeedback({ announcement: "Link copied" })
    expect(feedback.copied.value).toBe(false)

    await feedback.copy("https://example.com/e/1")

    expect(writeText).toHaveBeenCalledWith("https://example.com/e/1")
    expect(feedback.copied.value).toBe(true)
    expect(feedback.announcement.value).toBe("Link copied")

    vi.advanceTimersByTime(COPY_FEEDBACK_DURATION_MS - 1)
    expect(feedback.copied.value).toBe(true)

    vi.advanceTimersByTime(1)
    expect(feedback.copied.value).toBe(false)
    expect(feedback.announcement.value).toBe("")
  })

  it("restarts the revert timer when the user activates copy again", async () => {
    const feedback = useCopyFeedback()
    await feedback.copy("first")
    vi.advanceTimersByTime(1500)

    await feedback.copy("second")
    vi.advanceTimersByTime(1500)
    expect(feedback.copied.value).toBe(true)

    vi.advanceTimersByTime(500)
    expect(feedback.copied.value).toBe(false)
  })

  it("does not confirm or announce when the clipboard write fails", async () => {
    writeText.mockRejectedValue(new Error("denied"))
    const feedback = useCopyFeedback({ announcement: "Link copied" })

    await expect(feedback.copy("value")).rejects.toThrow("denied")

    expect(feedback.copied.value).toBe(false)
    expect(feedback.announcement.value).toBe("")
  })

  it("defaults the announcement to Copied", async () => {
    const feedback = useCopyFeedback()

    await feedback.copy("value")

    expect(feedback.announcement.value).toBe("Copied")
  })

  it("clears the pending timer when its scope is disposed", async () => {
    const scope = effectScope()
    const feedback = scope.run(() => useCopyFeedback())
    if (!feedback) throw new Error("Expected copy feedback state")
    await feedback.copy("value")
    expect(feedback.copied.value).toBe(true)

    scope.stop()

    expect(feedback.copied.value).toBe(false)
    vi.advanceTimersByTime(COPY_FEEDBACK_DURATION_MS)
    expect(feedback.copied.value).toBe(false)
  })

  it("resets copied state and the timer on demand", async () => {
    const feedback = useCopyFeedback()
    await feedback.copy("value")

    feedback.reset()

    expect(feedback.copied.value).toBe(false)
    vi.advanceTimersByTime(COPY_FEEDBACK_DURATION_MS)
    expect(feedback.copied.value).toBe(false)
  })
})
