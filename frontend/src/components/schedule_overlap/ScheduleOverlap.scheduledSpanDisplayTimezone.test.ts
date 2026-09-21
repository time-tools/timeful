// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from "vitest"
import { nextTick } from "vue"
import { Temporal } from "temporal-polyfill"
import { durations, eventTypes, UTC } from "@/constants"
import type * as EventTransportBoundary from "@/composables/event/eventTransportBoundary"
import type { ScheduleOverlapEvent } from "@/composables/schedule_overlap/types"
import { joinTooltipSegments } from "./scheduleOverlapRendering"
import { resetScheduleOverlapMocks } from "./scheduleOverlapTestMocks"
import {
  buildScheduleOverlapProps,
  getTimedGridPresentation,
  installScheduleOverlapTestGlobals,
  mountScheduleOverlap,
  zdt,
} from "./scheduleOverlapTestUtils"

const { saveTimefulScheduleMock } = vi.hoisted(() => ({
  saveTimefulScheduleMock: vi.fn(),
}))

vi.mock(
  "@/composables/event/eventTransportBoundary",
  async (importOriginal) => ({
    ...(await importOriginal<typeof EventTransportBoundary>()),
    saveTimefulSchedule: saveTimefulScheduleMock,
  }),
)

const displayTimezone = (value: string, offsetHours: number) => ({
  value,
  offset: Temporal.Duration.from({ hours: offsetHours }),
  label: value,
  gmtString: `GMT${offsetHours >= 0 ? "+" : ""}${String(offsetHours)}`,
})

const buildUtcWorkdayEvent = (): ScheduleOverlapEvent => {
  const selectedDay = Temporal.PlainDate.from("2026-06-02")
  const activeSlots = Array.from({ length: 32 }, (_, index) =>
    selectedDay
      .toZonedDateTime({ timeZone: UTC, plainTime: "09:00" })
      .add({ minutes: index * 15 }),
  )

  return {
    ...buildScheduleOverlapProps().event,
    _id: "evt-span-tz",
    shortId: "spantz",
    name: "Span timezone repro",
    type: eventTypes.SPECIFIC_DATES,
    dates: [selectedDay],
    timeSeed: zdt("2026-06-02T09:00:00Z"),
    startTime: Temporal.PlainTime.from("09:00"),
    duration: Temporal.Duration.from({ hours: 8 }),
    daysOnly: false,
    hasSpecificTimes: false,
    eventTimezone: UTC,
    slotGeneration: {
      startTimeLocal: Temporal.PlainTime.from("09:00"),
      endTimeLocal: Temporal.PlainTime.from("17:00"),
      timeIncrement: durations.FIFTEEN_MINUTES,
    },
    timedRecurrence: {
      kind: "specific_dates",
      selectedDays: [selectedDay],
      selectedDaysOfWeek: [],
      startOnMonday: true,
    },
    timeIncrement: durations.FIFTEEN_MINUTES,
    activeSlots,
  }
}

const buildUtcSlots = (
  selectedDay: Temporal.PlainDate,
  startTime: string,
  endTime: string,
) => {
  const start = selectedDay.toZonedDateTime({
    timeZone: UTC,
    plainTime: startTime,
  })
  const end = selectedDay.toZonedDateTime({ timeZone: UTC, plainTime: endTime })
  const slots: Temporal.ZonedDateTime[] = []
  for (
    let slot = start;
    Temporal.ZonedDateTime.compare(slot, end) < 0;
    slot = slot.add({ minutes: 15 })
  ) {
    slots.push(slot)
  }
  return slots
}

const buildGappedWorkdayEvent = (): ScheduleOverlapEvent => {
  const selectedDay = Temporal.PlainDate.from("2026-06-02")

  return {
    ...buildUtcWorkdayEvent(),
    _id: "evt-span-gap",
    shortId: "spangap",
    name: "Span collapse repro",
    activeSlots: [
      ...buildUtcSlots(selectedDay, "09:00", "10:00"),
      ...buildUtcSlots(selectedDay, "13:00", "17:00"),
    ],
  }
}

type SpanTimezoneVm = {
  scheduleEvent: () => void
  setScheduledEventFromRowCol: (scheduledEvent: {
    row: number
    col: number
    numRows: number
  }) => void
  curTimezone: ReturnType<typeof displayTimezone>
  tooltipContent: Parameters<typeof joinTooltipSegments>[0]
  getTimeslotVon: (row: number, col: number) => Record<string, () => void>
  confirmScheduleEvent: (destination: "timeful") => Promise<void>
}

const tooltipText = (vm: SpanTimezoneVm) =>
  joinTooltipSegments(vm.tooltipContent)

describe("ScheduleOverlap scheduled span display timezone", () => {
  beforeEach(() => {
    resetScheduleOverlapMocks()
    installScheduleOverlapTestGlobals()
    saveTimefulScheduleMock.mockReset()
    saveTimefulScheduleMock.mockResolvedValue(undefined)
  })

  it("keeps a pending span on its instants and follows the re-projected slots", async () => {
    const wrapper = mountScheduleOverlap({
      props: {
        event: buildUtcWorkdayEvent(),
        initialTimezone: displayTimezone("Etc/GMT-8", 8),
      },
    })
    const vm = wrapper.vm as unknown as SpanTimezoneVm

    vm.scheduleEvent()
    await nextTick()
    vm.setScheduledEventFromRowCol({ row: 72, col: 0, numRows: 8 })
    await nextTick()

    expect(getTimedGridPresentation(wrapper).curScheduledEvent).toEqual({
      row: 72,
      col: 0,
      numRows: 8,
    })

    vm.curTimezone = displayTimezone("Etc/GMT-7", 7)
    await nextTick()

    const presentation = getTimedGridPresentation(wrapper)
    expect(presentation.curScheduledEvent).toEqual({
      row: 68,
      col: 0,
      numRows: 8,
    })
    const renderedStartRow = presentation.renderedRows.find(
      (row) => row.baseRowIndex === 68,
    )
    if (!renderedStartRow) {
      throw new Error("Expected a rendered row at base row 68")
    }
    expect(presentation.scheduledEventStyles).toEqual([
      {
        top: `${String(renderedStartRow.rowTop)}px`,
        height: `${String(renderedStartRow.height * 8)}px`,
      },
    ])

    vm.getTimeslotVon(68, 0).mouseover()
    await nextTick()
    expect(tooltipText(vm)).toBe("5:00 PM to 7:00 PM \u00b7 Tue, Jun 2, 2026")

    await vm.confirmScheduleEvent("timeful")
    expect(saveTimefulScheduleMock).toHaveBeenCalledWith("spantz", {
      startDate: zdt("2026-06-02T10:00:00Z"),
      endDate: zdt("2026-06-02T12:00:00Z"),
    })

    wrapper.unmount()
  })

  it("keeps a saved span on its instants when the display timezone changes", async () => {
    const event = buildUtcWorkdayEvent()
    event.scheduledEvent = {
      startDate: zdt("2026-06-02T10:00:00Z"),
      endDate: zdt("2026-06-02T12:00:00Z"),
    }
    const wrapper = mountScheduleOverlap({
      props: {
        event,
        initialTimezone: displayTimezone("Etc/GMT-8", 8),
      },
    })
    const vm = wrapper.vm as unknown as SpanTimezoneVm

    expect(getTimedGridPresentation(wrapper).savedScheduledEvent).toEqual({
      row: 72,
      col: 0,
      numRows: 8,
    })

    vm.curTimezone = displayTimezone("Etc/GMT-7", 7)
    await nextTick()

    expect(getTimedGridPresentation(wrapper).savedScheduledEvent).toEqual({
      row: 68,
      col: 0,
      numRows: 8,
    })

    wrapper.unmount()
  })

  it("splits a pending span into fragments around a collapsed disabled run", async () => {
    const wrapper = mountScheduleOverlap({
      props: {
        event: buildGappedWorkdayEvent(),
        initialTimezone: displayTimezone("UTC", 0),
      },
    })
    const vm = wrapper.vm as unknown as SpanTimezoneVm

    vm.scheduleEvent()
    await nextTick()
    vm.setScheduledEventFromRowCol({ row: 38, col: 0, numRows: 17 })
    await nextTick()

    const presentation = getTimedGridPresentation(wrapper)
    expect(presentation.curScheduledEvent).toEqual({
      row: 38,
      col: 0,
      numRows: 17,
    })

    const firstRow = presentation.renderedRows.find(
      (row) => row.baseRowIndex === 38,
    )
    const secondRow = presentation.renderedRows.find(
      (row) => row.baseRowIndex === 52,
    )
    if (!firstRow || !secondRow) {
      throw new Error("Expected rendered rows around the collapsed run")
    }
    expect(presentation.scheduledEventStyles).toEqual([
      {
        top: `${String(firstRow.rowTop)}px`,
        height: `${String(firstRow.height * 2)}px`,
      },
      {
        top: `${String(secondRow.rowTop)}px`,
        height: `${String(secondRow.height * 3)}px`,
      },
    ])

    wrapper.unmount()
  })

  it("renders a pending span from a scheduled_event URL coordinate on mount", async () => {
    window.history.replaceState(
      {},
      "",
      `/?scheduled_event=${encodeURIComponent(
        JSON.stringify({ row: 72, col: 0, numRows: 8 }),
      )}`,
    )

    const wrapper = mountScheduleOverlap({
      props: {
        event: buildUtcWorkdayEvent(),
        initialTimezone: displayTimezone("Etc/GMT-8", 8),
      },
    })
    await nextTick()

    expect(getTimedGridPresentation(wrapper).curScheduledEvent).toEqual({
      row: 72,
      col: 0,
      numRows: 8,
    })
    expect(window.location.search).toBe("")

    wrapper.unmount()
    window.history.replaceState({}, "", "/")
  })
})
