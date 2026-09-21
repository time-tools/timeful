// @vitest-environment happy-dom

import { beforeEach, describe, expect, it } from "vitest"
import { nextTick } from "vue"
import { Temporal } from "temporal-polyfill"
import { joinTooltipSegments } from "./scheduleOverlapRendering"
import { resetScheduleOverlapMocks } from "./scheduleOverlapTestMocks"
import {
  buildScheduleOverlapProps,
  buildUtcSpecificTimes,
  installScheduleOverlapTestGlobals,
  mountScheduleOverlap,
  utcTimezone,
  zdt,
} from "./scheduleOverlapTestUtils"

const buildHourlyEvent = () => ({
  ...buildScheduleOverlapProps().event,
  dates: [Temporal.PlainDate.from("2026-01-01")],
  timeSeed: zdt("2026-01-01T09:00:00Z"),
  hasSpecificTimes: true,
  startTime: Temporal.PlainTime.from("09:00"),
  duration: Temporal.Duration.from({ hours: 4 }),
  timeIncrement: Temporal.Duration.from({ hours: 1 }),
  times: buildUtcSpecificTimes("2026-01-01", [
    "09:00:00",
    "10:00:00",
    "11:00:00",
    "12:00:00",
  ]),
})

const mountHourlyScheduling = () =>
  mountScheduleOverlap({
    props: {
      calendarOnly: true,
      event: buildHourlyEvent(),
      initialTimezone: utcTimezone,
    },
  })

type SchedulingTooltipVm = {
  state: string
  tooltipContent: Parameters<typeof joinTooltipSegments>[0]
  scheduleEvent: () => void
  cancelScheduleEvent: () => void
  curScheduledEvent: { row: number; col: number; numRows: number } | null
  dragging: boolean
  dragStart: { row: number; col: number } | null
  dragCur: { row: number; col: number } | null
  getTimeslotVon: (row: number, col: number) => Record<string, () => void>
}

const tooltipText = (vm: SchedulingTooltipVm) =>
  joinTooltipSegments(vm.tooltipContent)

describe("ScheduleOverlap scheduling tooltip", () => {
  beforeEach(() => {
    resetScheduleOverlapMocks()
    installScheduleOverlapTestGlobals()
  })

  it("reports the pending Timed Event Occurrence Span instead of the hovered Time Slot", async () => {
    const wrapper = mountHourlyScheduling()
    const vm = wrapper.vm as unknown as SchedulingTooltipVm

    vm.scheduleEvent()
    vm.curScheduledEvent = { row: 0, col: 0, numRows: 3 }
    await nextTick()

    vm.getTimeslotVon(1, 0).mouseover()
    await nextTick()

    expect(tooltipText(vm)).toBe("9:00 AM to 12:00 PM \u00b7 Thu, Jan 1, 2026")

    wrapper.unmount()
  })

  it("updates the reported span live while the schedule drag changes", async () => {
    const wrapper = mountHourlyScheduling()
    const vm = wrapper.vm as unknown as SchedulingTooltipVm

    vm.scheduleEvent()
    vm.dragging = true
    vm.dragStart = { row: 0, col: 0 }
    vm.dragCur = { row: 1, col: 0 }
    await nextTick()

    vm.getTimeslotVon(1, 0).mouseover()
    await nextTick()

    expect(tooltipText(vm)).toBe("9:00 AM to 11:00 AM \u00b7 Thu, Jan 1, 2026")

    vm.dragCur = { row: 2, col: 0 }
    await nextTick()

    vm.getTimeslotVon(1, 0).mouseover()
    await nextTick()

    expect(tooltipText(vm)).toBe("9:00 AM to 12:00 PM \u00b7 Thu, Jan 1, 2026")

    wrapper.unmount()
  })

  it("keeps reporting the pending span after the drag is released", async () => {
    const wrapper = mountHourlyScheduling()
    const vm = wrapper.vm as unknown as SchedulingTooltipVm

    vm.scheduleEvent()
    vm.curScheduledEvent = { row: 0, col: 0, numRows: 3 }
    await nextTick()

    vm.getTimeslotVon(2, 0).mouseover()
    await nextTick()

    expect(tooltipText(vm)).toBe("9:00 AM to 12:00 PM \u00b7 Thu, Jan 1, 2026")

    wrapper.unmount()
  })

  it("returns to the hovered Time Slot after scheduling is cancelled", async () => {
    const wrapper = mountHourlyScheduling()
    const vm = wrapper.vm as unknown as SchedulingTooltipVm

    vm.scheduleEvent()
    vm.curScheduledEvent = { row: 0, col: 0, numRows: 3 }
    await nextTick()

    vm.cancelScheduleEvent()
    await nextTick()

    vm.getTimeslotVon(1, 0).mouseover()
    await nextTick()

    expect(tooltipText(vm)).toBe("10:00 AM to 11:00 AM \u00b7 Thu, Jan 1, 2026")

    wrapper.unmount()
  })

  it("keeps Availability Editing hovering on the hovered Time Slot", async () => {
    const wrapper = mountHourlyScheduling()
    const vm = wrapper.vm as unknown as SchedulingTooltipVm

    vm.getTimeslotVon(1, 0).mouseover()
    await nextTick()

    expect(tooltipText(vm)).toBe("10:00 AM to 11:00 AM \u00b7 Thu, Jan 1, 2026")

    wrapper.unmount()
  })
})
