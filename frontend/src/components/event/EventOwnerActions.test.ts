// @vitest-environment happy-dom

import { flushPromises, mount } from "@vue/test-utils"
import { beforeEach, describe, expect, it, vi } from "vitest"
import MdiArchiveArrowUpOutline from "~icons/mdi/archive-arrow-up-outline"
import { clickButtonStub, passThroughStub } from "@/test/componentStubs"
import EventOwnerActions from "./EventOwnerActions.vue"

const { archiveEventMock, showErrorMock } = vi.hoisted(() => ({
  archiveEventMock: vi.fn(),
  showErrorMock: vi.fn(),
}))

vi.mock("@/utils/services/EventService", () => ({
  archiveEvent: archiveEventMock,
}))

vi.mock("@/stores/main", () => ({
  useMainStore: () => ({
    showError: showErrorMock,
  }),
}))

const baseEvent = (overrides: Record<string, unknown> = {}) => ({
  _id: "evt-1",
  name: "Managed event",
  eventVisitorId: "visitor-1",
  canManageEvent: true,
  isArchived: true,
  ...overrides,
})

const mountActions = (event: ReturnType<typeof baseEvent>) =>
  mount(EventOwnerActions, {
    props: { event },
    global: { stubs: { "v-btn": clickButtonStub, "v-icon": passThroughStub } },
  })

describe("EventOwnerActions", () => {
  beforeEach(() => {
    archiveEventMock.mockReset()
    showErrorMock.mockReset()
    archiveEventMock.mockResolvedValue(undefined)
  })

  it("offers Unarchive only for archived events the viewer can manage", () => {
    const archived = mountActions(baseEvent())
    expect(archived.find("button").text()).toBe("Unarchive event")

    expect(
      mountActions(baseEvent({ canManageEvent: false }))
        .find("button")
        .exists(),
    ).toBe(false)
    expect(
      mountActions(baseEvent({ isArchived: false }))
        .find("button")
        .exists(),
    ).toBe(false)
    expect(
      mountActions(baseEvent({ eventVisitorId: undefined }))
        .find("button")
        .exists(),
    ).toBe(false)
  })

  it("renders the unarchive icon to the left of the label", () => {
    const wrapper = mountActions(baseEvent())
    const button = wrapper.get("button")
    const icon = button.findComponent(MdiArchiveArrowUpOutline)

    expect(icon.exists()).toBe(true)
    const children = Array.from(button.element.children)
    expect(children[0].contains(icon.element)).toBe(true)
    expect(children[1].textContent).toBe("Unarchive event")
  })

  it("unarchives the event and requests a refresh", async () => {
    const wrapper = mountActions(baseEvent())

    await wrapper.get("button").trigger("click")
    await flushPromises()

    expect(archiveEventMock).toHaveBeenCalledWith("evt-1", false)
    expect(wrapper.emitted("changed")).toEqual([[]])
  })

  it("reports unarchive failures without emitting a refresh", async () => {
    archiveEventMock.mockRejectedValueOnce(new Error("nope"))
    const wrapper = mountActions(baseEvent())

    await wrapper.get("button").trigger("click")
    await flushPromises()

    expect(showErrorMock).toHaveBeenCalledWith(
      "Could not update the event. Refresh the page and try again.",
    )
    expect(wrapper.emitted("changed")).toBeUndefined()
  })
})
