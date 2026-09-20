// @vitest-environment happy-dom

import { computed, ref } from "vue"
import { describe, expect, it, vi } from "vitest"
import { availabilityTypes, type AvailabilityType } from "@/constants"
import { COMPACT_RESPONDENTS_PANEL_WIDTH } from "@/components/schedule_overlap/layout"
import { states, type ScheduleOverlapState } from "./types"
import { useScheduleOverlapUI } from "./useScheduleOverlapUI"
import { SCHEDULE_OVERLAP_COMPACT_DESKTOP_BREAKPOINT } from "@/components/schedule_overlap/scheduleOverlapBreakpoints"

vi.mock("@/stores/main", () => ({
  useMainStore: () => ({
    authUser: null,
  }),
}))

vi.mock("@/plugins/posthog", () => ({
  posthog: {
    capture: vi.fn(),
  },
}))

function createUi() {
  const isPhone = ref(false)
  const isSignUp = ref(false)
  const isGroup = ref(false)
  const daysOnly = ref(false)
  const state = ref<ScheduleOverlapState>(states.EDIT_AVAILABILITY)
  const availabilityType = ref<AvailabilityType>(availabilityTypes.AVAILABLE)
  const curTimeslot = ref({ row: 2, col: 3 })
  const curTimeslotAvailability = ref<Record<string, boolean>>({
    "user-1": true,
  })
  const curTimeslotInactive = ref(false)
  const curTimeslotCollapsed = ref(false)
  const timeslotSelected = ref(false)
  const endDrag = vi.fn()

  const ui = useScheduleOverlapUI({
    isPhone,
    isSignUp: computed(() => isSignUp.value),
    isGroup: computed(() => isGroup.value),
    daysOnly,
    state,
    showBestTimes: ref(false),
    defaultState: computed(() => states.HEATMAP),
    allowDrag: computed(() => true),
    availabilityType,
    parsedResponses: computed(() => ({})),
    curTimeslot,
    endDrag,
    timeslotSelected,
    curTimeslotAvailability,
    curTimeslotInactive,
    curTimeslotCollapsed,
    respondents: computed(() => [{ _id: "user-1" }]),
    curGuestId: ref(""),
    guestName: computed(() => undefined),
    ownedGuestResponseLookupKeys: computed(() => new Set<string>()),
    guestResponseLookupKey: computed(() => undefined),
    guestAddedAvailability: computed(() => false),
  })

  return {
    ui,
    isPhone,
    isSignUp,
    isGroup,
    daysOnly,
    state,
    availabilityType,
    curTimeslot,
    curTimeslotInactive,
    curTimeslotCollapsed,
    timeslotSelected,
    endDrag,
  }
}

describe("useScheduleOverlapUI deselectRespondents", () => {
  it("keeps the current timeslot when the release click lands inside the drag section", () => {
    const { ui, curTimeslot, endDrag } = createUi()
    const dragSection = document.createElement("div")
    dragSection.id = "drag-section"
    const inner = document.createElement("div")
    dragSection.appendChild(inner)
    document.body.appendChild(dragSection)

    const event = new MouseEvent("click", { bubbles: true })
    Object.defineProperty(event, "target", {
      configurable: true,
      value: inner,
    })
    ui.deselectRespondents(event)

    expect(curTimeslot.value).toEqual({ row: 2, col: 3 })
    expect(endDrag).not.toHaveBeenCalled()

    dragSection.remove()
  })

  it("keeps the current timeslot when a click lands inside the mobile overlay", () => {
    const { ui, curTimeslot, endDrag } = createUi()
    const mobileOverlay = document.createElement("div")
    mobileOverlay.className = "schedule-overlap-mobile-overlay"
    const inner = document.createElement("div")
    mobileOverlay.appendChild(inner)
    document.body.appendChild(mobileOverlay)

    const event = new MouseEvent("click", { bubbles: true })
    Object.defineProperty(event, "target", {
      configurable: true,
      value: inner,
    })
    ui.deselectRespondents(event)

    expect(curTimeslot.value).toEqual({ row: 2, col: 3 })
    expect(endDrag).not.toHaveBeenCalled()

    mobileOverlay.remove()
  })

  it("clears the current timeslot when a click lands outside the grid and overlay", () => {
    const { ui, curTimeslot, endDrag } = createUi()
    const outside = document.createElement("button")
    document.body.appendChild(outside)

    const event = new MouseEvent("click", { bubbles: true })
    Object.defineProperty(event, "target", {
      configurable: true,
      value: outside,
    })
    ui.deselectRespondents(event)

    expect(curTimeslot.value).toEqual({ row: -1, col: -1 })
    expect(endDrag).toHaveBeenCalledTimes(1)

    outside.remove()
  })

  it("clears the inactive-slot flag when resetting the current timeslot", () => {
    const { ui, curTimeslotInactive } = createUi()
    curTimeslotInactive.value = true

    ui.resetCurTimeslot(true)

    expect(curTimeslotInactive.value).toBe(false)
  })

  it("clears the collapsed-hours flag when resetting the current timeslot", () => {
    const { ui, curTimeslotCollapsed } = createUi()
    curTimeslotCollapsed.value = true

    ui.resetCurTimeslot(true)

    expect(curTimeslotCollapsed.value).toBe(false)
  })

  it("keeps a mobile timeslot selected when the grid receives mouseleave", () => {
    const { ui, isPhone, curTimeslot, endDrag } = createUi()
    isPhone.value = true

    ui.resetCurTimeslot()

    expect(curTimeslot.value).toEqual({ row: 2, col: 3 })
    expect(endDrag).not.toHaveBeenCalled()
  })

  it("clears the timeslot when deselecting respondent selection even on mobile", () => {
    const {
      ui,
      isPhone,
      curTimeslot,
      curTimeslotInactive,
      timeslotSelected,
      endDrag,
    } = createUi()
    isPhone.value = true
    timeslotSelected.value = true
    curTimeslotInactive.value = true

    ui.deselectRespondentsSelection()

    expect(curTimeslot.value).toEqual({ row: -1, col: -1 })
    expect(timeslotSelected.value).toBe(false)
    expect(curTimeslotInactive.value).toBe(false)
    expect(endDrag).toHaveBeenCalledTimes(1)
  })

  it("uses a fixed respondents panel width on compact desktop", () => {
    const { ui } = createUi()

    expect(ui.rightSideWidth.value).toBe(COMPACT_RESPONDENTS_PANEL_WIDTH)
  })

  it("keeps full-width mobile and fixed sign-up respondents widths", () => {
    const { ui, isPhone, isSignUp } = createUi()

    isPhone.value = true
    expect(ui.rightSideWidth.value).toBe("100%")

    isPhone.value = false
    isSignUp.value = true
    expect(ui.rightSideWidth.value).toBe("18rem")
  })

  it("uses the shared 640px breakpoint contract for compact desktop layout decisions", () => {
    expect(SCHEDULE_OVERLAP_COMPACT_DESKTOP_BREAKPOINT).toBe(640)
  })
})

describe("useScheduleOverlapUI hintText", () => {
  it("points at the grid below and adapts the verb to the viewport", () => {
    const { ui, isPhone } = createUi()

    expect(ui.hintText.value).toBe(
      'Click and drag on the grid below to add your "available" times in green.',
    )

    isPhone.value = true
    expect(ui.hintText.value).toBe(
      'Tap and drag on the grid below to add your "available" times in green.',
    )
  })

  it("describes if-needed availability in yellow", () => {
    const { ui, availabilityType } = createUi()
    availabilityType.value = availabilityTypes.IF_NEEDED

    expect(ui.hintText.value).toBe(
      'Click and drag on the grid below to add your "if needed" times in yellow.',
    )
  })

  it("describes toggling calendars for availability groups", () => {
    const { ui, isGroup } = createUi()
    isGroup.value = true

    expect(ui.hintText.value).toBe(
      "Toggle which calendars are used. Click and drag on the grid below to edit your availability.",
    )
  })

  it("points at the grid below while scheduling", () => {
    const { ui, state } = createUi()
    state.value = states.SCHEDULE_EVENT

    expect(ui.hintText.value).toBe(
      "Click and drag on the grid below to schedule a Google Calendar event during those times.",
    )
  })

  it("uses days for dates-only events", () => {
    const { ui, daysOnly } = createUi()
    daysOnly.value = true

    expect(ui.hintText.value).toBe(
      'Click and drag on the grid below to add your "available" days in green.',
    )
  })

  it("describes if-needed dates-only availability in yellow", () => {
    const { ui, daysOnly, availabilityType } = createUi()
    daysOnly.value = true
    availabilityType.value = availabilityTypes.IF_NEEDED

    expect(ui.hintText.value).toBe(
      'Click and drag on the grid below to add your "if needed" days in yellow.',
    )
  })

  it("uses days while scheduling a dates-only event", () => {
    const { ui, daysOnly, state } = createUi()
    daysOnly.value = true
    state.value = states.SCHEDULE_EVENT

    expect(ui.hintText.value).toBe(
      "Click and drag on the grid below to schedule a Google Calendar event during those days.",
    )
  })

  it("does not produce a hint outside editing and scheduling", () => {
    const { ui, state } = createUi()
    state.value = states.HEATMAP

    expect(ui.hintText.value).toBe("")
  })
})
