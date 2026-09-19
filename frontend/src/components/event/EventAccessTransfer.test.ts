// @vitest-environment happy-dom
import { flushPromises, mount } from "@vue/test-utils"
import { afterEach, beforeEach, expect, it, vi } from "vitest"
import EventAccessTransfer from "./EventAccessTransfer.vue"
import { eventTypes } from "@/constants"
import { createLocalStorageMock } from "@/test/localStorage"

vi.mock("@/stores/main", () => ({ useMainStore: () => ({}) }))
import { savedTransfers } from "@/composables/transfer/transferBoundary"
import { FetchError } from "@/utils/fetch_utils"
import type * as FetchUtils from "@/utils/fetch_utils"

const post = vi.hoisted(() => vi.fn())
vi.mock("@/utils/fetch_utils", async (original) => ({
  ...(await original<typeof FetchUtils>()),
  post,
}))
beforeEach(() => {
  vi.useFakeTimers()
  vi.stubGlobal("localStorage", createLocalStorageMock())
  post
    .mockReset()
    .mockImplementation((url: string) =>
      Promise.resolve(
        url.endsWith("/transfers")
          ? { id: "transfer", state: "pending" }
          : { state: "pending", requests: [] },
      ),
    )
})
afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})
function render() {
  return mount(EventAccessTransfer, {
    props: {
      event: {
        _id: "EVENT123",
        eventVisitorId: "visitor",
        name: "Event",
        type: eventTypes.SPECIFIC_DATES,
      },
    },
    global: {
      stubs: {
        VBtn: { template: "<button><slot /></button>" },
        VDialog: {
          props: { contentProps: Object, scrollable: Boolean },
          template:
            '<div role="dialog" :aria-labelledby="contentProps && contentProps[\'aria-labelledby\']" :data-scrollable="String(scrollable)"><slot /></div>',
        },
        VCard: { template: "<div><slot /></div>" },
        VCardTitle: { template: "<div><slot /></div>" },
        VCardText: { template: "<div><slot /></div>" },
        VCardActions: { template: "<div><slot /></div>" },
        VAlert: { template: '<div role="alert"><slot /></div>' },
        VTextField: {
          props: ["label", "modelValue"],
          emits: ["update:modelValue"],
          template:
            '<input :aria-label="label" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
        },
      },
    },
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
function steps(wrapper: ReturnType<typeof render>) {
  return wrapper.findAll('[data-testid^="manage-access-step-"]')
}
it("renders the trigger with the green outlined treatment", () => {
  const wrapper = render()
  const button = wrapper
    .findAll("button")
    .find((button) => button.text() === "Manage access")

  expect(button).toBeDefined()
  expect(button?.attributes("variant")).toBe("outlined")
  expect(button?.attributes("color")).toBe("primary")
  expect(button?.get("span").classes()).toContain("tw:text-green")
})

it("explains the transfer flow in the dialog", () => {
  const text = render().text()

  expect(text).toContain("Use this event on another browser")
  expect(text).toContain("revoke access you granted earlier")
  expect(text).toContain("The link expires five minutes after you create it")
  expect(text).toContain("It shows a matching code")
  expect(text).toContain("Opening the link alone gives no access")
  expect(text).toContain("Enter its code and approve")
})

it("renders the three numbered steps behind an accessible dialog name", () => {
  const wrapper = render()

  expect(wrapper.get('[role="dialog"]').attributes("aria-labelledby")).toBe(
    "manage-access-title",
  )
  expect(wrapper.get("#manage-access-title").text()).toBe("Manage access")
  expect(wrapper.get('[role="dialog"]').attributes("data-scrollable")).toBe(
    "true",
  )
  const rendered = steps(wrapper)
  expect(rendered.map((step) => step.attributes("data-testid"))).toEqual([
    "manage-access-step-1",
    "manage-access-step-2",
    "manage-access-step-3",
  ])
  expect(rendered[0].text()).toContain(
    "Step 1: Create and copy a transfer link",
  )
  expect(rendered[1].text()).toContain(
    "Step 2: Open the link in the other browser",
  )
  expect(rendered[2].text()).toContain("Step 3: Enter its code and approve")
  wrapper.unmount()
})

it("starts with step 1 current and hides the code input until a transfer exists", () => {
  const wrapper = render()
  const rendered = steps(wrapper)

  expect(rendered[0].attributes("aria-current")).toBe("step")
  expect(rendered[1].attributes("aria-current")).toBeUndefined()
  expect(rendered[2].attributes("aria-current")).toBeUndefined()
  expect(
    wrapper
      .find('input[aria-label="Matching code from other browser"]')
      .exists(),
  ).toBe(false)
  wrapper.unmount()
})

it("marks step 2 current after the link is copied and waits for the other browser", async () => {
  vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue(undefined)
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  await click(wrapper, "Copy link")

  expect(steps(wrapper)[1].attributes("aria-current")).toBe("step")
  expect(wrapper.text()).toContain(
    "Waiting for the other browser to open the link.",
  )
  wrapper.unmount()
})

it("marks step 3 current when the other browser is showing a code", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockImplementation((url: string) =>
    Promise.resolve(
      url.endsWith("/transfers")
        ? { id: "transfer", state: "pending" }
        : {
            state: "pending",
            requests: [{ id: "request", code: "ABC123" }],
          },
    ),
  )
  await vi.advanceTimersByTimeAsync(2100)
  await flushPromises()

  expect(steps(wrapper)[2].attributes("aria-current")).toBe("step")
  expect(wrapper.text()).toContain(
    "The other browser is showing a code — enter it in step 3.",
  )
  wrapper.unmount()
})

it("keeps step 3 current while approved and returns to step 1 when terminal", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockImplementation((url: string) =>
    Promise.resolve(
      url.endsWith("/transfers")
        ? { id: "transfer", state: "pending" }
        : { state: "approved", requests: [{ id: "request", code: "ABC123" }] },
    ),
  )
  await vi.advanceTimersByTimeAsync(2100)
  await flushPromises()

  expect(steps(wrapper)[2].attributes("aria-current")).toBe("step")
  expect(wrapper.text()).toContain("Approved — finish in the other browser.")

  post.mockResolvedValue({ state: "redeemed", requests: [], revocable: false })
  await vi.advanceTimersByTimeAsync(2100)
  await flushPromises()

  expect(steps(wrapper)[0].attributes("aria-current")).toBe("step")
  expect(wrapper.text()).toContain("Completed.")
  wrapper.unmount()
})

it("keeps the code input available before the first status poll and describes it", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  const input = wrapper.get(
    'input[aria-label="Matching code from other browser"]',
  )

  expect(input.attributes("autocapitalize")).toBe("characters")
  expect(input.attributes("spellcheck")).toBe("false")
  expect(input.attributes("hint")).toContain("other browser")
  const approve = wrapper
    .findAll("button")
    .find((button) => button.text() === "Approve matching code")
  expect(approve?.attributes("disabled")).toBeDefined()

  await input.setValue("ABC123")

  expect(approve?.attributes("disabled")).toBeUndefined()
  wrapper.unmount()
})

it("announces the copied link politely", async () => {
  vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue(undefined)
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  await click(wrapper, "Copy link")

  const announcements = wrapper
    .findAll('[aria-live="polite"]')
    .map((status) => status.text())
  expect(announcements).toContain("Transfer link copied")
  wrapper.unmount()
})

it("restores a saved pending transfer after a reload", async () => {
  localStorage.setItem(
    "timeful.transfers.EVENT123",
    JSON.stringify({
      nextNumber: 2,
      transfers: [{ id: "transfer", number: 1 }],
    }),
  )
  const wrapper = render()
  await click(wrapper, "Manage access")

  expect(wrapper.text()).toContain("Copy link")
  expect(wrapper.text()).toContain(
    "Waiting for the other browser to open the link.",
  )
  expect(wrapper.text()).toContain("Approve matching code")
  expect(wrapper.text()).toContain("Cancel transfer")
  expect(
    wrapper.get('input[aria-label="Transfer link"]').attributes("value"),
  ).toBe("http://localhost:3000/transfer/EVENT123/transfer")
  expect(
    wrapper
      .get('input[aria-label="Matching code from other browser"]')
      .attributes("aria-label"),
  ).toBe("Matching code from other browser")
  wrapper.unmount()
})

it("restores a saved approved transfer after a reload", async () => {
  localStorage.setItem(
    "timeful.transfers.EVENT123",
    JSON.stringify({
      nextNumber: 2,
      transfers: [{ id: "transfer", number: 1 }],
    }),
  )
  post.mockImplementation((url: string) =>
    Promise.resolve(
      url.endsWith("/transfers")
        ? { id: "transfer", state: "pending" }
        : { state: "approved", requests: [] },
    ),
  )
  const wrapper = render()
  await click(wrapper, "Manage access")

  expect(wrapper.text()).toContain("Approved — finish in the other browser.")
  expect(wrapper.text()).toContain("Cancel transfer")
  expect(wrapper.text()).not.toContain("Approve matching code")
  expect(wrapper.text()).toContain("Copy link")
  wrapper.unmount()
})

it("does not restore a terminal saved transfer as current after a reload", async () => {
  localStorage.setItem(
    "timeful.transfers.EVENT123",
    JSON.stringify({
      nextNumber: 2,
      transfers: [{ id: "transfer", number: 1 }],
    }),
  )
  post.mockImplementation((url: string) =>
    Promise.resolve(
      url.endsWith("/transfers")
        ? { id: "transfer", state: "pending" }
        : { state: "expired", requests: [] },
    ),
  )
  const wrapper = render()
  await click(wrapper, "Manage access")

  expect(wrapper.text()).toContain("Create new transfer link")
  expect(wrapper.text()).not.toContain("Copy link")
  expect(wrapper.text()).not.toContain("Cancel transfer")
  expect(wrapper.find('input[aria-label="Transfer link"]').exists()).toBe(false)
  wrapper.unmount()
})

it("restores the most recently created active transfer after a reload", async () => {
  localStorage.setItem(
    "timeful.transfers.EVENT123",
    JSON.stringify({
      nextNumber: 3,
      transfers: [
        { id: "first", number: 1 },
        { id: "second", number: 2 },
      ],
    }),
  )
  post.mockImplementation((url: string) =>
    Promise.resolve(
      url.endsWith("/transfers")
        ? { id: "second", state: "pending" }
        : { state: url.includes("/first/") ? "pending" : "approved" },
    ),
  )
  const wrapper = render()
  await click(wrapper, "Manage access")

  expect(
    wrapper.get('input[aria-label="Transfer link"]').attributes("value"),
  ).toBe("http://localhost:3000/transfer/EVENT123/second")
  expect(wrapper.text()).toContain("Approved — finish in the other browser.")
  expect(wrapper.text()).toContain("Cancel transfer")
  wrapper.unmount()
})

it("cancels the live transfer before creating a replacement link", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockImplementation((url: string) =>
    Promise.resolve(
      url.endsWith("/cancel")
        ? { state: "cancelled" }
        : { id: "replacement", state: "pending" },
    ),
  )

  await click(wrapper, "Create new transfer link")

  expect(post).toHaveBeenCalledWith(
    "/events/EVENT123/transfers/transfer/cancel",
    {},
  )
  expect(post).toHaveBeenLastCalledWith("/events/EVENT123/transfers")
  expect(
    wrapper.get('input[aria-label="Transfer link"]').attributes("value"),
  ).toBe("http://localhost:3000/transfer/EVENT123/replacement")
  expect(savedTransfers("EVENT123").map(({ id }) => id)).toEqual([
    "replacement",
  ])
  expect(wrapper.text()).toContain(
    "Waiting for the other browser to open the link.",
  )
  wrapper.unmount()
})

it("keeps the live transfer visible while its replacement is created", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  let resolveCancel: (value: { state: string }) => void = () => {}
  let resolveCreate: (value: { id: string; state: string }) => void = () => {}
  post.mockImplementation((url: string) =>
    url.endsWith("/cancel")
      ? new Promise((resolve) => {
          resolveCancel = resolve
        })
      : new Promise((resolve) => {
          resolveCreate = resolve
        }),
  )

  const button = wrapper
    .findAll("button")
    .find((button) => button.text() === "Create new transfer link")
  await button?.trigger("click")
  await flushPromises()
  resolveCancel({ state: "cancelled" })
  await flushPromises()

  expect(wrapper.text()).toContain(
    "Waiting for the other browser to open the link.",
  )
  expect(wrapper.text()).toContain("Copy link")
  expect(wrapper.text()).toContain("Approve matching code")
  expect(wrapper.text()).toContain("Cancel transfer")
  expect(
    wrapper.get('input[aria-label="Transfer link"]').attributes("value"),
  ).toBe("http://localhost:3000/transfer/EVENT123/transfer")

  resolveCreate({ id: "replacement", state: "pending" })
  await flushPromises()

  expect(
    wrapper.get('input[aria-label="Transfer link"]').attributes("value"),
  ).toBe("http://localhost:3000/transfer/EVENT123/replacement")
  expect(wrapper.text()).toContain(
    "Waiting for the other browser to open the link.",
  )
  wrapper.unmount()
})

it("reconciles a cancelled transfer when the replacement create fails", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockImplementation((url: string) =>
    url.endsWith("/cancel")
      ? Promise.resolve({ state: "cancelled" })
      : Promise.reject(
          Object.assign(new FetchError("Server error"), { status: 500 }),
        ),
  )

  await click(wrapper, "Create new transfer link")

  expect(wrapper.get('[role="alert"]').text()).toContain(
    "Could not create a transfer link",
  )
  expect(wrapper.text()).toContain("Cancelled.")
  expect(wrapper.text()).not.toContain("Copy link")
  expect(savedTransfers("EVENT123")).toEqual([])
  wrapper.unmount()
})

it("keeps the other transfer actions enabled while a replacement create is pending", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  await wrapper
    .get('input[aria-label="Matching code from other browser"]')
    .setValue("ABC123")
  post.mockImplementation(() => new Promise(() => {}))
  const findButton = (text: string) =>
    wrapper.findAll("button").find((button) => button.text() === text)

  await findButton("Create new transfer link")?.trigger("click")
  await flushPromises()

  expect(
    findButton("Create new transfer link")?.attributes("disabled"),
  ).toBeDefined()
  expect(findButton("Cancel transfer")?.attributes("disabled")).toBeUndefined()
  expect(
    findButton("Approve matching code")?.attributes("disabled"),
  ).toBeUndefined()
  wrapper.unmount()
})

it("keeps Create new transfer link enabled while a copy is pending", async () => {
  let resolveWrite: () => void = () => {}
  vi.spyOn(navigator.clipboard, "writeText").mockImplementation(
    () =>
      new Promise<void>((resolve) => {
        resolveWrite = resolve
      }),
  )
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  const createButton = wrapper
    .findAll("button")
    .find((button) => button.text() === "Create new transfer link")

  await click(wrapper, "Copy link")

  expect(createButton?.attributes("disabled")).toBeUndefined()
  expect(wrapper.text()).not.toContain("Copied")

  resolveWrite()
  await flushPromises()

  expect(wrapper.text()).toContain("Copied")
  wrapper.unmount()
})

it("copies the current link without creating another transfer", async () => {
  const writeText = vi
    .spyOn(navigator.clipboard, "writeText")
    .mockResolvedValue(undefined)
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockClear()

  await click(wrapper, "Copy link")

  expect(post).not.toHaveBeenCalled()
  expect(writeText).toHaveBeenCalledWith(
    "http://localhost:3000/transfer/EVENT123/transfer",
  )
  expect(wrapper.text()).toContain("Copied")
  expect(savedTransfers("EVENT123")).toHaveLength(1)
  wrapper.unmount()
})

it("creates a new transfer without cancelling after a terminal transfer", async () => {
  localStorage.setItem(
    "timeful.transfers.EVENT123",
    JSON.stringify({
      nextNumber: 2,
      transfers: [{ id: "transfer", number: 1 }],
    }),
  )
  post.mockImplementation((url: string) =>
    Promise.resolve(
      url.endsWith("/transfers")
        ? { id: "fresh", state: "pending" }
        : { state: "expired", requests: [] },
    ),
  )
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")

  expect(post).not.toHaveBeenCalledWith(
    "/events/EVENT123/transfers/transfer/cancel",
    {},
  )
  expect(
    wrapper.get('input[aria-label="Transfer link"]').attributes("value"),
  ).toBe("http://localhost:3000/transfer/EVENT123/fresh")
  expect(savedTransfers("EVENT123").map(({ id }) => id)).toEqual(["fresh"])
  wrapper.unmount()
})

it("hides Copy link once the transfer is no longer live", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  expect(wrapper.text()).toContain("Copy link")
  post.mockResolvedValue({ state: "cancelled" })

  await click(wrapper, "Cancel transfer")

  expect(wrapper.text()).not.toContain("Copy link")
  expect(wrapper.text()).toContain("Cancelled.")
  expect(wrapper.find('input[aria-label="Transfer link"]').exists()).toBe(true)
  wrapper.unmount()
})

it("keeps wrong-code feedback visible across status polling", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  await wrapper
    .get('input[aria-label="Matching code from other browser"]')
    .setValue("WRONG")
  await click(wrapper, "Approve matching code")
  expect(wrapper.get('[role="alert"]').text()).toContain("Check the code")
  await vi.advanceTimersByTimeAsync(2100)
  await flushPromises()
  expect(wrapper.get('[role="alert"]').text()).toContain("Check the code")
  wrapper.unmount()
})

it("prunes legacy history and polls only active transfers with stable grant numbers", async () => {
  localStorage.setItem(
    "timeful.transfers.EVENT123",
    JSON.stringify(["old", "first", "second", "session", "missing", "first"]),
  )
  let revoked = false
  post.mockImplementation((url: string) => {
    if (url.includes("/missing/"))
      return Promise.reject(
        Object.assign(new FetchError("Forbidden"), { status: 403 }),
      )
    if (url.endsWith("/revoke")) {
      revoked = true
      return Promise.resolve({ state: "revoked" })
    }
    if (url.includes("/first/"))
      return Promise.resolve({
        state: revoked ? "cancelled" : "redeemed",
        revocable: !revoked,
      })
    if (url.includes("/second/"))
      return Promise.resolve({ state: "redeemed", revocable: true })
    return Promise.resolve({
      state: url.includes("/session/") ? "redeemed" : "expired",
    })
  })
  const wrapper = render()
  await click(wrapper, "Manage access")
  expect(post).toHaveBeenCalledTimes(5)
  expect(savedTransfers("EVENT123").map(({ id }) => id)).toEqual([
    "first",
    "second",
  ])
  expect(wrapper.text()).toContain("Granted access 3")
  post.mockClear()
  await vi.advanceTimersByTimeAsync(2000)
  expect(post).toHaveBeenCalledTimes(2)
  await click(wrapper, "Revoke access")
  expect(wrapper.text()).not.toContain("Granted access 2")
  expect(wrapper.text()).toContain("Granted access 3")
  post.mockClear()
  await vi.advanceTimersByTimeAsync(2000)
  expect(post).toHaveBeenCalledTimes(1)
  await click(wrapper, "Close")
  post.mockClear()
  await vi.advanceTimersByTimeAsync(6000)
  expect(post).not.toHaveBeenCalled()
  wrapper.unmount()
  const reopened = render()
  await click(reopened, "Manage access")
  expect(reopened.text()).toContain("Granted access 3")
  reopened.unmount()
  post.mockClear()
  await vi.advanceTimersByTimeAsync(4000)
  expect(post).not.toHaveBeenCalled()
})

it("polls the current transfer once and stops after a non-revocable completion", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockClear()
  await vi.advanceTimersByTimeAsync(2000)
  expect(post).toHaveBeenCalledTimes(1)
  expect(savedTransfers("EVENT123")).toHaveLength(1)
  post.mockResolvedValue({ state: "approved" })
  await vi.advanceTimersByTimeAsync(2000)
  expect(savedTransfers("EVENT123")).toHaveLength(1)
  expect(wrapper.text()).toContain("Approved — finish in the other browser.")
  post.mockResolvedValue({ state: "redeemed", revocable: false })
  await vi.advanceTimersByTimeAsync(2000)
  expect(savedTransfers("EVENT123")).toEqual([])
  expect(wrapper.text()).toContain("Completed.")
  post.mockClear()
  await vi.advanceTimersByTimeAsync(4000)
  expect(post).not.toHaveBeenCalled()
  wrapper.unmount()
})

it("allows cancelling an approved transfer and stops tracking it", async () => {
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockResolvedValue({ state: "approved" })
  await vi.advanceTimersByTimeAsync(2000)
  expect(wrapper.text()).toContain("Approved — finish in the other browser.")
  expect(wrapper.text()).not.toContain("Approve matching code")
  post.mockResolvedValue({ state: "cancelled" })
  await click(wrapper, "Cancel transfer")
  expect(post).toHaveBeenLastCalledWith(
    "/events/EVENT123/transfers/transfer/cancel",
    {},
  )
  expect(wrapper.text()).toContain("Cancelled.")
  expect(wrapper.text()).not.toContain("Cancel transfer")
  expect(savedTransfers("EVENT123")).toEqual([])
  post.mockClear()
  await vi.advanceTimersByTimeAsync(4000)
  expect(post).not.toHaveBeenCalled()
  wrapper.unmount()
})

it("does not drop a cancel while a status poll is in flight", async () => {
  let resolveStatus: (value: { state: string }) => void = () => {}
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  post.mockResolvedValue({ state: "approved" })
  await vi.advanceTimersByTimeAsync(2000)
  expect(wrapper.text()).toContain("Approved — finish in the other browser.")
  post.mockImplementation((url: string) =>
    url.endsWith("/cancel")
      ? Promise.resolve({ state: "cancelled" })
      : new Promise((resolve) => {
          resolveStatus = resolve
        }),
  )
  await vi.advanceTimersByTimeAsync(2000)
  await click(wrapper, "Cancel transfer")
  expect(post).toHaveBeenLastCalledWith(
    "/events/EVENT123/transfers/transfer/cancel",
    {},
  )
  resolveStatus({ state: "approved" })
  await flushPromises()
  expect(wrapper.text()).toContain("Cancelled.")
  expect(wrapper.text()).not.toContain("Cancel transfer")
  expect(savedTransfers("EVENT123")).toEqual([])
  wrapper.unmount()
})

it("retains revocation handles and visible grants on transient status failures", async () => {
  localStorage.setItem("timeful.transfers.EVENT123", JSON.stringify(["grant"]))
  post.mockResolvedValue({ state: "redeemed", revocable: true })
  const wrapper = render()
  await click(wrapper, "Manage access")
  post.mockRejectedValue(
    Object.assign(new FetchError("Server error"), { status: 500 }),
  )
  await vi.advanceTimersByTimeAsync(2000)
  expect(savedTransfers("EVENT123")).toHaveLength(1)
  expect(wrapper.text()).toContain("Granted access 1")
  post.mockResolvedValue({ state: "redeemed", revocable: true })
  post.mockClear()
  await vi.advanceTimersByTimeAsync(2000)
  expect(post).toHaveBeenCalledTimes(1)
  wrapper.unmount()
})

it("names clipboard failures and preserves the message across successful polls", async () => {
  vi.spyOn(navigator.clipboard, "writeText").mockRejectedValue(
    new Error("denied"),
  )
  const wrapper = render()
  await click(wrapper, "Manage access")
  await click(wrapper, "Create new transfer link")
  await click(wrapper, "Copy link")
  expect(wrapper.get('[role="alert"]').text()).toContain(
    "Could not copy the transfer link",
  )
  expect(wrapper.get('[role="alert"]').text()).not.toContain("expired")
  await vi.advanceTimersByTimeAsync(2000)
  expect(wrapper.get('[role="alert"]').text()).toContain(
    "Could not copy the transfer link",
  )
  wrapper.unmount()
})
