// @vitest-environment happy-dom

import { computed, defineComponent, nextTick, ref } from "vue"
import { mount } from "@vue/test-utils"
import { describe, expect, it } from "vitest"
import { Temporal } from "temporal-polyfill"
import {
  availabilityTypes,
  eventTypes,
  timeTypes,
  type TimeType,
  UTC,
} from "@/constants"
import {
  states,
  type ScheduleOverlapEvent,
  type TimeItem,
} from "@/composables/schedule_overlap/types"
import { useTimedGridPresentation } from "./useTimedGridPresentation"

const time = (hour: number): TimeItem => ({
  hoursOffset: Temporal.Duration.from({ hours: hour }),
  absoluteMinutes: hour * 60,
})

const mountPresentation = () => {
  const collapseDisabledTimesPreference = ref(true)
  const splitTimes = ref<TimeItem[][]>([
    [time(9), time(10), time(11), time(12), time(13)],
    [],
  ])
  const timeType = ref<TimeType>(timeTypes.HOUR24)
  const times = computed(() => splitTimes.value.flat())
  let presentation!: ReturnType<typeof useTimedGridPresentation>

  const Harness = defineComponent({
    setup() {
      presentation = useTimedGridPresentation({
        event: computed<ScheduleOverlapEvent>(() => ({
          _id: "evt-1",
          name: "Presentation test event",
          type: eventTypes.SPECIFIC_DATES,
          dates: [Temporal.PlainDate.from("2026-01-01")],
          timeSeed: Temporal.Instant.from(
            "2026-01-01T09:00:00Z",
          ).toZonedDateTimeISO(UTC),
          daysOnly: false,
        })),
        state: ref(states.HEATMAP),
        defaultState: computed(() => states.HEATMAP),
        // This isolates row projection from availability-color calculations.
        isSignUp: computed(() => true),
        collapseDisabledTimesPreference,
        availabilityType: ref(availabilityTypes.AVAILABLE),
        curGuestId: computed(() => ""),
        authUserId: computed(() => undefined),
        animateTimeslotAlways: computed(() => false),
        availabilityAnimEnabled: ref(false),
        curRespondentsMax: computed(() => 0),
        dragging: ref(false),
        dragStart: ref(null),
        dragCur: ref(null),
        schedulingGridPointerVisible: computed(() => false),
        getTimeslotVon: () => ({}),
        grid: {
          splitTimes,
          timeType,
          timeslotDuration: ref(Temporal.Duration.from({ hours: 1 })),
          days: ref([{}]),
          times,
          timeslotHeight: ref(40),
          getDateFromRowCol: () => null,
          getEnabledDateFromRowCol: (row: number) =>
            Temporal.Instant.from(
              `2026-01-01T${String(row + 9).padStart(2, "0")}:00:00Z`,
            ).toZonedDateTimeISO(UTC),
        } as never,
        avail: {} as never,
        drag: {} as never,
        scheduling: {} as never,
        ui: {} as never,
      })
      return () => null
    },
  })

  const wrapper = mount(Harness)
  return {
    presentation,
    collapseDisabledTimesPreference,
    timeType,
    splitTimes,
    wrapper,
  }
}

describe("useTimedGridPresentation", () => {
  it("projects collapsed rows and restores their base rows when expanded", () => {
    const { presentation, wrapper } = mountPresentation()

    expect(presentation.renderedRows.value).toEqual([
      expect.objectContaining({
        id: "collapsed-540-840",
        kind: "collapsed",
        startLabel: "09:00",
        endLabel: "14:00",
      }),
    ])
    expect(presentation.timeAxisEndText.value).toBe("14:00")

    presentation.toggleCollapsedSpan("collapsed-540-840")

    expect(
      presentation.renderedRows.value.map((row) => row.baseRowIndex),
    ).toEqual([0, 1, 2, 3, 4])
    wrapper.unmount()
  })

  it("clears expanded spans and renders the full axis when expanded", () => {
    const { presentation, collapseDisabledTimesPreference, wrapper } =
      mountPresentation()

    presentation.updateCollapseDisabledTimes(false)

    expect(collapseDisabledTimesPreference.value).toBe(false)
    expect(presentation.collapseDisabledTimes.value).toBe(false)
    expect(presentation.renderedRows.value).toHaveLength(5)
    wrapper.unmount()
  })

  it("derives the switch state from manual expansions and collapse-all", () => {
    const { presentation, collapseDisabledTimesPreference, wrapper } =
      mountPresentation()

    expect(presentation.collapseDisabledTimes.value).toBe(true)

    presentation.toggleCollapsedSpan("collapsed-540-840")
    expect(presentation.collapseDisabledTimes.value).toBe(false)
    expect(collapseDisabledTimesPreference.value).toBe(true)
    expect(presentation.renderedRows.value).toHaveLength(5)

    presentation.updateCollapseDisabledTimes(true)
    expect(presentation.collapseDisabledTimes.value).toBe(true)
    expect(collapseDisabledTimesPreference.value).toBe(true)
    expect(presentation.renderedRows.value).toEqual([
      expect.objectContaining({ id: "collapsed-540-840", kind: "collapsed" }),
    ])
    wrapper.unmount()
  })

  it("keeps manually expanded hours open when timezone projection changes their labels", async () => {
    const { presentation, splitTimes, wrapper } = mountPresentation()

    presentation.toggleCollapsedSpan("collapsed-540-840")
    expect(presentation.renderedRows.value).toHaveLength(5)

    splitTimes.value = [[time(4), time(5), time(6), time(7), time(8)], []]
    await nextTick()

    expect(presentation.renderedRows.value.map((row) => row.kind)).toEqual([
      "timeslot",
      "timeslot",
      "timeslot",
      "timeslot",
      "timeslot",
    ])
    wrapper.unmount()
  })

  it("updates all grid labels when the time format changes", () => {
    const { presentation, timeType, wrapper } = mountPresentation()

    timeType.value = timeTypes.HOUR12

    expect(presentation.renderedRows.value).toEqual([
      expect.objectContaining({ startLabel: "9 AM", endLabel: "2 PM" }),
    ])
    expect(presentation.timeAxisEndText.value).toBe("2 PM")
    wrapper.unmount()
  })
})
