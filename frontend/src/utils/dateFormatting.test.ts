import { describe, expect, it } from "vitest"
import { Temporal } from "temporal-polyfill"

import { eventTypes, timeTypes, UTC } from "@/constants"
import {
  getDateRangeString,
  getDateRangeStringForEvent,
  getDaysInMonth,
  getEventOccurrenceSpanString,
  getISODateString,
  getStartEndDateString,
  timeNumToTimeString,
  timeNumToTimeText,
} from "./dateFormatting"

describe("dateFormatting", () => {
  const zdt = (iso: string) =>
    Temporal.Instant.from(iso).toZonedDateTimeISO(UTC)

  it("formats midnight-ended ranges as inclusive of the prior day", () => {
    expect(
      getDateRangeString(
        zdt("2026-05-01T00:00:00Z"),
        zdt("2026-05-03T00:00:00Z"),
        true,
      ),
    ).toBe("5/1 - 5/2")
  })

  it("formats event date ranges for weekly and specific-date events", () => {
    expect(
      getDateRangeStringForEvent({
        type: eventTypes.DOW,
        dates: [
          Temporal.PlainDate.from("2026-05-03"),
          Temporal.PlainDate.from("2026-05-04"),
        ],
      }),
    ).toBe("Sun, Mon")

    expect(
      getDateRangeStringForEvent({
        type: eventTypes.SPECIFIC_DATES,
        dates: [
          Temporal.PlainDate.from("2026-05-01"),
          Temporal.PlainDate.from("2026-05-03"),
        ],
        timeSeed: zdt("2026-05-01T00:00:00Z"),
      }),
    ).toBe("5/1 - 5/3")
  })

  it("derives timed specific-date summaries from picked dates", () => {
    expect(
      getDateRangeStringForEvent({
        type: eventTypes.SPECIFIC_DATES,
        dates: [
          Temporal.PlainDate.from("2026-05-28"),
          Temporal.PlainDate.from("2026-05-29"),
        ],
        timeSeed: zdt("2026-05-28T00:00:00Z"),
        timedRecurrence: {
          kind: "specific_dates",
          selectedDays: [
            Temporal.PlainDate.from("2026-05-28"),
            Temporal.PlainDate.from("2026-05-29"),
          ],
          selectedDaysOfWeek: [],
          startOnMonday: true,
        },
        eventTimezone: UTC,
        slotGeneration: {
          startTimeLocal: Temporal.PlainTime.from("00:00:00"),
          endTimeLocal: Temporal.PlainTime.from("02:00:00"),
          timeIncrement: Temporal.Duration.from({ minutes: 60 }),
        },
      }),
    ).toBe("5/28 - 5/29")
  })

  it("keeps timed specific-date summaries on the picked dates in the persisted event timezone", () => {
    expect(
      getDateRangeStringForEvent({
        type: eventTypes.SPECIFIC_DATES,
        dates: [
          Temporal.PlainDate.from("2026-01-05"),
          Temporal.PlainDate.from("2026-01-06"),
        ],
        timedRecurrence: {
          kind: "specific_dates",
          selectedDays: [
            Temporal.PlainDate.from("2026-01-05"),
            Temporal.PlainDate.from("2026-01-06"),
          ],
          selectedDaysOfWeek: [],
          startOnMonday: true,
        },
        eventTimezone: "America/Los_Angeles",
        slotGeneration: {
          startTimeLocal: Temporal.PlainTime.from("23:30:00"),
          endTimeLocal: Temporal.PlainTime.from("01:30:00"),
          timeIncrement: Temporal.Duration.from({ minutes: 30 }),
        },
      }),
    ).toBe("1/5 - 1/6")
  })

  it("can format timed specific-date summaries in the viewer timezone", () => {
    expect(
      getDateRangeStringForEvent(
        {
          type: eventTypes.SPECIFIC_DATES,
          dates: [
            Temporal.PlainDate.from("2026-06-11"),
            Temporal.PlainDate.from("2026-06-12"),
          ],
          eventTimezone: "Asia/Seoul",
          slotGeneration: {
            startTimeLocal: Temporal.PlainTime.from("09:00:00"),
            endTimeLocal: Temporal.PlainTime.from("17:00:00"),
            timeIncrement: Temporal.Duration.from({ minutes: 15 }),
          },
        },
        {
          value: "America/Los_Angeles",
          offset: Temporal.Duration.from({ hours: 7 }),
          label: "America/Los_Angeles",
          gmtString: "GMT-7",
        },
      ),
    ).toBe("6/10 - 6/11")
  })

  it("formats time numbers for display and transport", () => {
    expect(timeNumToTimeText(0)).toBe("12 AM")
    expect(timeNumToTimeText(13.5)).toBe("1:30 PM")
    expect(timeNumToTimeText(13.5, false)).toBe("13:30")
    expect(timeNumToTimeText(9)).toBe("9 AM")
    expect(timeNumToTimeString(9.5)).toBe("09:30:00")
  })

  it("keeps calendar formatting helpers on Temporal-native values", () => {
    expect(getISODateString(zdt("2026-05-01T12:00:00Z"), true)).toBe(
      "2026-05-01",
    )
    expect(
      getStartEndDateString(
        zdt("2026-05-01T09:00:00Z"),
        zdt("2026-05-01T10:30:00Z"),
      ),
    ).toContain("Fri, May 1")
    expect(getDaysInMonth(2, 2028)).toBe(29)
  })

  const utcTimezone = {
    value: UTC,
    offset: Temporal.Duration.from({ minutes: 0 }),
    label: "UTC",
    gmtString: "GMT+00:00",
  }

  it("formats a same-date Timed Event Occurrence Span with the display time format", () => {
    const startDate = zdt("2026-09-15T08:00:00Z")
    const endDate = zdt("2026-09-15T08:15:00Z")

    expect(
      getEventOccurrenceSpanString({
        startDate,
        endDate,
        timezone: utcTimezone,
        timeType: timeTypes.HOUR12,
        daysOnly: false,
      }),
    ).toBe("Tue, Sep 15, 2026 \u00b7 8:00 AM \u2013 8:15 AM")

    expect(
      getEventOccurrenceSpanString({
        startDate,
        endDate,
        timezone: utcTimezone,
        timeType: timeTypes.HOUR24,
        daysOnly: false,
      }),
    ).toBe("Tue, Sep 15, 2026 \u00b7 08:00 \u2013 08:15")
  })

  it("reports both Civil Dates when a timed occurrence span crosses midnight", () => {
    expect(
      getEventOccurrenceSpanString({
        startDate: zdt("2026-07-04T23:30:00Z"),
        endDate: zdt("2026-07-05T00:30:00Z"),
        timezone: utcTimezone,
        timeType: timeTypes.HOUR12,
        daysOnly: false,
      }),
    ).toBe(
      "Sat, Jul 4, 2026 \u00b7 11:30 PM \u2013 Sun, Jul 5, 2026 \u00b7 12:30 AM",
    )
  })

  it("formats a timed occurrence span in the Display Timezone", () => {
    expect(
      getEventOccurrenceSpanString({
        startDate: zdt("2026-09-15T15:00:00Z"),
        endDate: zdt("2026-09-15T15:30:00Z"),
        timezone: {
          value: "America/Los_Angeles",
          offset: Temporal.Duration.from({ hours: -7 }),
          label: "America/Los_Angeles",
          gmtString: "GMT-7",
        },
        timeType: timeTypes.HOUR12,
        daysOnly: false,
      }),
    ).toBe("Tue, Sep 15, 2026 \u00b7 8:00 AM \u2013 8:30 AM")
  })

  it("formats a timed occurrence span in a fixed-offset Display Timezone", () => {
    expect(
      getEventOccurrenceSpanString({
        startDate: zdt("2026-09-15T08:00:00Z"),
        endDate: zdt("2026-09-15T08:30:00Z"),
        timezone: {
          value: "",
          offset: Temporal.Duration.from({ hours: 5, minutes: 30 }),
          label: "GMT+5:30",
          gmtString: "GMT+5:30",
        },
        timeType: timeTypes.HOUR12,
        daysOnly: false,
      }),
    ).toBe("Tue, Sep 15, 2026 \u00b7 1:30 PM \u2013 2:00 PM")
  })

  it("formats a Dates-Only Event Occurrence Span as a single date without a time range", () => {
    expect(
      getEventOccurrenceSpanString({
        startDate: zdt("2026-05-28T00:00:00Z"),
        endDate: zdt("2026-05-29T00:00:00Z"),
        timezone: utcTimezone,
        timeType: timeTypes.HOUR12,
        daysOnly: true,
      }),
    ).toBe("Thu, May 28, 2026")
  })
})
