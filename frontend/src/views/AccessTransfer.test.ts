// @vitest-environment happy-dom
import { flushPromises, mount } from "@vue/test-utils"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { FetchError } from "@/utils/fetch_utils"
import AccessTransfer from "./AccessTransfer.vue"
import type * as TransferBoundary from "@/composables/transfer/transferBoundary"

const mocks = vi.hoisted(() => ({ action: vi.fn(), navigate: vi.fn() }))
vi.mock("@/stores/main", () => ({
  useMainStore: () => ({
    authUser: {
      firstName: "Current",
      lastName: "User",
      email: "current@example.com",
    },
  }),
}))
vi.mock("@/composables/transfer/transferBoundary", async (original) => ({
  ...(await original<typeof TransferBoundary>()),
  transferAction: mocks.action,
}))
const stubs = {
  VContainer: { template: "<div><slot /></div>" },
  VCard: { props: ["title"], template: "<div>{{ title }}<slot /></div>" },
  VCardText: { template: "<div><slot /></div>" },
  VCardActions: { template: "<div><slot /></div>" },
  VAlert: { template: '<div role="alert"><slot /></div>' },
  VBtn: { template: "<button><slot /></button>" },
  VDialog: {
    props: ["modelValue"],
    template: '<div v-if="modelValue" role="dialog"><slot /></div>',
  },
}
function render() {
  return mount(AccessTransfer, {
    props: { eventId: "EVENT123", transferId: "transfer" },
    global: { stubs },
  })
}
async function click(wrapper: ReturnType<typeof render>, text: string) {
  const button = wrapper
    .findAll("button")
    .find((button) => button.text() === text)
  expect(button).toBeDefined()
  await button?.trigger("click")
  await flushPromises()
}
function switchRequired() {
  return Object.assign(new FetchError("Conflict"), {
    status: 409,
    parsed: { accountSwitchRequired: true },
  })
}
function httpError(status: number, message = "Request failed") {
  return Object.assign(new FetchError(message), { status })
}
beforeEach(() => {
  mocks.action
    .mockReset()
    .mockResolvedValue({ state: "approved", code: "12345678" })
  mocks.navigate.mockReset()
  vi.spyOn(window.location, "assign").mockImplementation(mocks.navigate)
})
afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})
describe("target access transfer", () => {
  it("presents the target side as numbered steps 2 and 3", async () => {
    const wrapper = render()
    await flushPromises()
    const step2 = wrapper.get('[data-testid="access-transfer-step-2"]')
    const step3 = wrapper.get('[data-testid="access-transfer-step-3"]')

    expect(step2.text()).toContain(
      "Step 2: Show this code to the browser that created the link",
    )
    expect(step2.text()).toContain("within five minutes")
    expect(step3.text()).toContain("Step 3: Continue after approval")
    expect(step2.attributes("aria-current")).toBeUndefined()
    expect(step3.attributes("aria-current")).toBe("step")
    wrapper.unmount()
  })
  it("keeps step 2 current and withheld access visible until approval", async () => {
    mocks.action.mockResolvedValueOnce({ state: "pending", code: "12345678" })
    const wrapper = render()
    await flushPromises()
    const step2 = wrapper.get('[data-testid="access-transfer-step-2"]')
    const step3 = wrapper.get('[data-testid="access-transfer-step-3"]')

    expect(step2.attributes("aria-current")).toBe("step")
    expect(step3.attributes("aria-current")).toBeUndefined()
    expect(step3.text()).toContain("no transferred access")
    expect(wrapper.text()).not.toContain("Approved — you can continue")
    wrapper.unmount()
  })
  it("restores the approved code on a fresh mount and can redeem", async () => {
    const wrapper = render()
    await flushPromises()
    expect(mocks.action).toHaveBeenCalledWith("EVENT123", "transfer", "open")
    expect(wrapper.get('[data-testid="matching-code"]').text()).toBe("12345678")
    expect(wrapper.text()).toContain("Approved — you can continue")
    expect(
      wrapper
        .get('[data-testid="access-transfer-step-3"]')
        .attributes("aria-current"),
    ).toBe("step")
    await click(wrapper, "Continue after approval")
    expect(mocks.navigate).toHaveBeenCalledWith("/e/EVENT123")
    wrapper.unmount()
  })
  it("requires explicit account-switch consent and cancellation does not redeem", async () => {
    const wrapper = render()
    await flushPromises()
    mocks.action.mockRejectedValueOnce(switchRequired())
    await click(wrapper, "Continue after approval")
    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.text()).toContain("Switch accounts on this device?")
    expect(dialog.text()).toContain("Current User")
    expect(dialog.text()).toContain("current@example.com")
    expect(dialog.text()).toContain("replacing your current sign-in")
    expect(dialog.text()).toContain("does not merge accounts")
    expect(mocks.navigate).not.toHaveBeenCalled()
    await click(wrapper, "Cancel")
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(mocks.action).toHaveBeenCalledTimes(2)
    mocks.action.mockRejectedValueOnce(switchRequired())
    await click(wrapper, "Continue after approval")
    await click(wrapper, "Switch accounts")
    expect(mocks.action).toHaveBeenLastCalledWith(
      "EVENT123",
      "transfer",
      "redeem",
      { confirmAccountSwitch: true },
    )
    expect(mocks.navigate).toHaveBeenCalledWith("/e/EVENT123")
    wrapper.unmount()
  })
  it("keeps unavailable-transfer errors visible without prompting a switch", async () => {
    const wrapper = render()
    await flushPromises()
    mocks.action.mockRejectedValueOnce(new FetchError("Forbidden"))
    await click(wrapper, "Continue after approval")
    expect(wrapper.text()).toContain("Access has not been approved")
    expect(wrapper.get('[role="alert"]').attributes("type")).toBe("error")
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(mocks.navigate).not.toHaveBeenCalled()
    wrapper.unmount()
  })
  it("shows the unavailable link message when opening fails", async () => {
    mocks.action.mockRejectedValueOnce(new FetchError("Forbidden"))
    const wrapper = render()
    await flushPromises()
    expect(wrapper.text()).toContain(
      "This transfer is expired, cancelled, or unavailable",
    )
    expect(wrapper.find("button").exists()).toBe(false)
    expect(wrapper.get('[role="alert"]').attributes("type")).toBe("error")
    wrapper.unmount()
  })
  it("highlights step 3 when the source approves without a reload", async () => {
    vi.useFakeTimers()
    mocks.action.mockResolvedValueOnce({ state: "pending", code: "12345678" })
    const wrapper = render()
    await flushPromises()
    expect(
      wrapper
        .get('[data-testid="access-transfer-step-3"]')
        .attributes("aria-current"),
    ).toBeUndefined()

    mocks.action.mockResolvedValue({ state: "approved", code: "12345678" })
    await vi.advanceTimersByTimeAsync(2100)
    await flushPromises()

    expect(
      wrapper
        .get('[data-testid="access-transfer-step-3"]')
        .attributes("aria-current"),
    ).toBe("step")
    expect(wrapper.get('[role="status"]').text()).toBe(
      "Approved — you can continue",
    )
    wrapper.unmount()
  })
  it("stops polling once the source approves", async () => {
    vi.useFakeTimers()
    mocks.action.mockResolvedValueOnce({ state: "pending", code: "12345678" })
    const wrapper = render()
    await flushPromises()

    mocks.action.mockResolvedValue({ state: "approved", code: "12345678" })
    await vi.advanceTimersByTimeAsync(2100)
    await flushPromises()
    const calls = mocks.action.mock.calls.length
    expect(calls).toBe(2)

    await vi.advanceTimersByTimeAsync(6000)
    await flushPromises()

    expect(mocks.action).toHaveBeenCalledTimes(calls)
    wrapper.unmount()
  })
  it("does not poll when the transfer is already approved", async () => {
    vi.useFakeTimers()
    const wrapper = render()
    await flushPromises()

    await vi.advanceTimersByTimeAsync(6000)
    await flushPromises()

    expect(mocks.action).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })
  it("keeps waiting when a status poll fails after the first open", async () => {
    vi.useFakeTimers()
    mocks.action.mockResolvedValueOnce({ state: "pending", code: "12345678" })
    const wrapper = render()
    await flushPromises()

    mocks.action.mockRejectedValue(new Error("network"))
    await vi.advanceTimersByTimeAsync(2100)
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('[role="status"]').text()).toContain(
      "Waiting for approval in the other browser",
    )
    wrapper.unmount()
  })
  it("keeps waiting and polling through a transient 500", async () => {
    vi.useFakeTimers()
    mocks.action.mockResolvedValueOnce({ state: "pending", code: "12345678" })
    const wrapper = render()
    await flushPromises()

    mocks.action.mockRejectedValue(httpError(500, "Server Error"))
    await vi.advanceTimersByTimeAsync(2100)
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('[role="status"]').text()).toContain(
      "Waiting for approval in the other browser",
    )
    expect(wrapper.find("button").exists()).toBe(true)

    const calls = mocks.action.mock.calls.length
    await vi.advanceTimersByTimeAsync(2100)
    await flushPromises()

    expect(mocks.action.mock.calls.length).toBeGreaterThan(calls)
    wrapper.unmount()
  })
  it.each([403, 404])(
    "ends the wait and stops polling when a poll definitively fails with %i",
    async (status) => {
      vi.useFakeTimers()
      mocks.action.mockResolvedValueOnce({ state: "pending", code: "12345678" })
      const wrapper = render()
      await flushPromises()

      mocks.action.mockRejectedValue(httpError(status))
      await vi.advanceTimersByTimeAsync(2100)
      await flushPromises()

      expect(wrapper.get('[role="alert"]').text()).toContain(
        "This transfer is expired, cancelled, or unavailable",
      )
      expect(wrapper.get('[role="status"]').text()).toContain(
        "This link is expired, cancelled, or unavailable",
      )
      expect(wrapper.find("button").exists()).toBe(false)

      const calls = mocks.action.mock.calls.length
      await vi.advanceTimersByTimeAsync(6000)
      await flushPromises()

      expect(mocks.action).toHaveBeenCalledTimes(calls)
      wrapper.unmount()
    },
  )
  it("clears an early redeem error when approval arrives", async () => {
    vi.useFakeTimers()
    mocks.action.mockResolvedValueOnce({ state: "pending", code: "12345678" })
    const wrapper = render()
    await flushPromises()

    mocks.action.mockRejectedValueOnce(new FetchError("Forbidden"))
    await click(wrapper, "Continue after approval")
    expect(wrapper.get('[role="alert"]').text()).toContain(
      "Access has not been approved",
    )

    mocks.action.mockResolvedValue({ state: "approved", code: "12345678" })
    await vi.advanceTimersByTimeAsync(2100)
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.get('[role="status"]').text()).toBe(
      "Approved — you can continue",
    )
    wrapper.unmount()
  })
})
