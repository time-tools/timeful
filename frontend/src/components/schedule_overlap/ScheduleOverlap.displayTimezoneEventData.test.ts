// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from "vitest"
import { nextTick } from "vue"
import { Temporal } from "temporal-polyfill"
import { durations, eventTypes, UTC } from "@/constants"
import type { ScheduleOverlapEvent } from "@/composables/schedule_overlap/types"
import { resetScheduleOverlapMocks } from "./scheduleOverlapTestMocks"
import {
  buildScheduleOverlapProps,
  installScheduleOverlapTestGlobals,
  mountScheduleOverlap,
  zdt,
} from "./scheduleOverlapTestUtils"

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
    _id: "evt-timezone-data",
    shortId: "tzdata",
    name: "Display timezone data stability",
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

type DisplayTimezoneVm = {
  curTimezone: ReturnType<typeof displayTimezone>
}

describe("ScheduleOverlap display timezone event data", () => {
  beforeEach(() => {
    resetScheduleOverlapMocks()
    installScheduleOverlapTestGlobals()
  })

  it("leaves Event Picked Dates and Active Slots unchanged and writes no event on a Display Timezone change", async () => {
    const putUrls: string[] = []
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
        const method =
          init?.method ?? (input instanceof Request ? input.method : "GET")
        const url = String(input instanceof Request ? input.url : input)
        if (method === "PUT") putUrls.push(url)
        return Promise.resolve({
          ok: true,
          status: 200,
          statusText: "OK",
          headers: new Headers(),
          text: () => Promise.resolve("{}"),
          json: () => Promise.resolve({}),
        } as Response)
      }),
    )

    const event = buildUtcWorkdayEvent()
    const wrapper = mountScheduleOverlap({
      props: {
        event,
        initialTimezone: displayTimezone("Etc/GMT-8", 8),
      },
    })
    const vm = wrapper.vm as unknown as DisplayTimezoneVm
    const eventProp = () => wrapper.props("event") as ScheduleOverlapEvent
    const originalDates = (eventProp().dates ?? []).map((date) =>
      date.toString(),
    )
    const originalActiveSlots = (eventProp().activeSlots ?? []).map((slot) =>
      slot.toInstant().toString(),
    )

    vm.curTimezone = displayTimezone("Etc/GMT-7", 7)
    await nextTick()

    expect((eventProp().dates ?? []).map((date) => date.toString())).toEqual(
      originalDates,
    )
    expect(
      (eventProp().activeSlots ?? []).map((slot) =>
        slot.toInstant().toString(),
      ),
    ).toEqual(originalActiveSlots)
    expect(putUrls).toEqual([])

    wrapper.unmount()
  })
})
