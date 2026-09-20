// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from "vitest"
import {
  resetScheduleOverlapMocks,
  viewportWidth,
} from "./scheduleOverlapTestMocks"
import {
  installScheduleOverlapTestGlobals,
  mountScheduleOverlap,
  type ScheduleOverlapWrapper,
} from "./scheduleOverlapTestUtils"

const { calendarAutofillEnabledState } = vi.hoisted(() => ({
  calendarAutofillEnabledState: { value: true },
}))

vi.mock("@/utils/calendarAutofillAvailability", () => ({
  get calendarAutofillEnabled() {
    return calendarAutofillEnabledState.value
  },
}))

const sidebarShowCalendarOptions = (wrapper: ScheduleOverlapWrapper) =>
  (
    wrapper
      .findComponent({ name: "ScheduleOverlapSidebar" })
      .props("sidebar") as {
      showCalendarOptions: boolean
    }
  ).showCalendarOptions

const mobileOverlayShowCalendarOptions = (wrapper: ScheduleOverlapWrapper) =>
  (
    wrapper
      .findComponent({ name: "ScheduleOverlapMobileOverlay" })
      .props("overlay") as { showCalendarOptions: boolean }
  ).showCalendarOptions

const mountWithViewModelStubs = () =>
  mountScheduleOverlap({
    global: {
      stubs: {
        ScheduleOverlapSidebar: {
          name: "ScheduleOverlapSidebar",
          props: { sidebar: { type: Object, required: true } },
          template: "<div />",
        },
        ScheduleOverlapMobileOverlay: {
          name: "ScheduleOverlapMobileOverlay",
          props: { overlay: { type: Object, required: true } },
          template: "<div />",
        },
      },
    },
  })

describe("ScheduleOverlap calendar options gating", () => {
  beforeEach(() => {
    resetScheduleOverlapMocks()
    installScheduleOverlapTestGlobals()
    calendarAutofillEnabledState.value = true
  })

  it("hides calendar options from the child view models when calendar autofill is disabled", () => {
    calendarAutofillEnabledState.value = false
    viewportWidth.value = 400

    const wrapper = mountWithViewModelStubs()

    expect(sidebarShowCalendarOptions(wrapper)).toBe(false)
    expect(mobileOverlayShowCalendarOptions(wrapper)).toBe(false)
  })

  it("shows calendar options in the child view models when calendar autofill is enabled without calendar permission", () => {
    viewportWidth.value = 400

    const wrapper = mountWithViewModelStubs()

    expect(sidebarShowCalendarOptions(wrapper)).toBe(true)
    expect(mobileOverlayShowCalendarOptions(wrapper)).toBe(true)
  })
})
