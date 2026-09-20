// @vitest-environment happy-dom

import { computed, nextTick, ref } from "vue"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { createLocalStorageMock } from "@/test/localStorage"
import type { ScheduleOverlapInstance } from "./types"
import { useEventLoader } from "./useEventLoader"

const {
  getMock,
  fetchEventFromPathMock,
  fetchCalendarEventsMapMock,
  calendarAutofillEnabledState,
} = vi.hoisted(() => ({
  getMock: vi.fn(),
  fetchEventFromPathMock: vi.fn(),
  fetchCalendarEventsMapMock: vi.fn(),
  calendarAutofillEnabledState: { value: true },
}))

vi.mock("@/utils/calendarAutofillAvailability", () => ({
  get calendarAutofillEnabled() {
    return calendarAutofillEnabledState.value
  },
}))

vi.mock("@/utils", () => ({
  get: getMock,
  getRenderedWeekStart: vi.fn(),
  processEvent: vi.fn(),
}))

vi.mock("./eventTransportBoundary", () => ({
  fetchEventFromPath: fetchEventFromPathMock,
}))

vi.mock("./calendarEventsBoundary", () => ({
  fetchCalendarEventsMap: fetchCalendarEventsMapMock,
  fetchCalendarAvailabilities: vi.fn(),
}))

describe("useEventLoader", () => {
  beforeEach(() => {
    vi.stubGlobal("localStorage", createLocalStorageMock())
    getMock.mockReset()
    fetchEventFromPathMock.mockReset()
    fetchCalendarEventsMapMock.mockReset()
    calendarAutofillEnabledState.value = true
    getMock.mockResolvedValue({ longId: "evt.long" })
    fetchEventFromPathMock.mockResolvedValue({
      _id: "evt.long",
      type: "specificDates",
      responses: {},
    })
    fetchCalendarEventsMapMock.mockResolvedValue({})
  })

  it("does not reuse whitespace-only stored guest names when loading events", async () => {
    globalThis.localStorage.setItem("evt.long.guestName", "   ")

    const loader = useEventLoader({
      eventId: ref("evt.long"),
      weekOffset: ref(0),
      authUser: computed(() => null),
    })

    await loader.refreshEvent()

    expect(fetchEventFromPathMock).toHaveBeenCalledWith("/events/evtlong")
  })

  it("reports calendar permission granted for a signed-in account with no calendars", async () => {
    const loader = useEventLoader({
      eventId: ref("evt.long"),
      weekOffset: ref(0),
      authUser: computed(() => ({ _id: "user-1" })),
    })

    await loader.refreshEvent()
    fetchCalendarEventsMapMock.mockResolvedValue({})
    await loader.fetchAuthUserCalendarEvents()

    expect(fetchCalendarEventsMapMock).toHaveBeenCalled()
    expect(loader.calendarPermissionGranted.value).toBe(true)
  })

  it("does not refill availability automatically when calendar autofill is disabled", async () => {
    calendarAutofillEnabledState.value = false
    const setAvailabilityAutomatically = vi.fn()
    const scheduleOverlapRef = ref<ScheduleOverlapInstance | null>({
      setAvailabilityAutomatically,
    } as unknown as ScheduleOverlapInstance)
    const loader = useEventLoader({
      eventId: ref("evt.long"),
      weekOffset: ref(0),
      authUser: computed(() => ({ _id: "user-1" })),
      scheduleOverlapRef,
      isEditing: computed(() => true),
      userHasResponded: computed(() => false),
      areUnsavedChanges: computed(() => false),
    })

    await loader.refreshEvent()
    fetchCalendarEventsMapMock.mockResolvedValue({})
    await loader.fetchAuthUserCalendarEvents()
    await nextTick()

    expect(setAvailabilityAutomatically).not.toHaveBeenCalled()
  })

  it("refills availability automatically after calendar events load when calendar autofill is enabled", async () => {
    const setAvailabilityAutomatically = vi.fn()
    const scheduleOverlapRef = ref<ScheduleOverlapInstance | null>({
      setAvailabilityAutomatically,
    } as unknown as ScheduleOverlapInstance)
    const loader = useEventLoader({
      eventId: ref("evt.long"),
      weekOffset: ref(0),
      authUser: computed(() => ({ _id: "user-1" })),
      scheduleOverlapRef,
      isEditing: computed(() => true),
      userHasResponded: computed(() => false),
      areUnsavedChanges: computed(() => false),
    })

    await loader.refreshEvent()
    fetchCalendarEventsMapMock.mockResolvedValue({})
    await loader.fetchAuthUserCalendarEvents()
    await nextTick()

    expect(setAvailabilityAutomatically).toHaveBeenCalledTimes(1)
  })
})
