import { describe, expect, it, vi } from "vitest"
import { Temporal } from "temporal-polyfill"
import { durations, eventTypes, UTC } from "@/constants"
import type * as UtilsModule from "@/utils"
import type { Event } from "@/types"
import { buildEventEditorSchedule } from "./eventEditorSchedule"
import {
  applySpecificTimesEditDraft,
  buildSpecificTimesCreateDraft,
  buildSpecificTimesEditDraft,
} from "./specificTimesEditDraft"

const { processEventMock } = vi.hoisted(() => ({
  processEventMock: vi.fn(),
}))

vi.mock("@/utils", async () => {
  const actual = await vi.importActual<typeof UtilsModule>("@/utils")
  return {
    ...actual,
    processEvent: processEventMock,
  }
})

const buildEvent = (): Event => ({
  _id: "evt-1",
  name: "Specific-times event",
  type: "specific_dates",
  dates: [
    Temporal.PlainDate.from("2026-05-30"),
    Temporal.PlainDate.from("2026-05-31"),
  ],
  timeSeed: Temporal.Instant.from("2026-05-30T09:00:00Z").toZonedDateTimeISO(
    UTC,
  ),
  duration: durations.ONE_HOUR,
  hasSpecificTimes: true,
  timeIncrement: durations.FIFTEEN_MINUTES,
  eventTimezone: UTC,
  slotGeneration: {
    startTimeLocal: Temporal.PlainTime.from("09:00"),
    endTimeLocal: Temporal.PlainTime.from("09:30"),
    timeIncrement: durations.FIFTEEN_MINUTES,
  },
  timedRecurrence: {
    kind: "specific_dates",
    selectedDays: [
      Temporal.PlainDate.from("2026-05-30"),
      Temporal.PlainDate.from("2026-05-31"),
    ],
    selectedDaysOfWeek: [],
    startOnMonday: true,
  },
  times: [
    Temporal.Instant.from("2026-05-30T00:00:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-05-31T00:00:00Z").toZonedDateTimeISO(UTC),
  ],
  activeSlots: [
    Temporal.Instant.from("2026-05-31T09:00:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-05-31T09:15:00Z").toZonedDateTimeISO(UTC),
  ],
})

const buildWeeklyEvent = (): Event => ({
  _id: "evt-weekly",
  name: "Weekly specific-times event",
  type: eventTypes.DOW,
  dates: [
    Temporal.PlainDate.from("2026-01-05"),
    Temporal.PlainDate.from("2026-01-07"),
  ],
  timeSeed: Temporal.Instant.from("2026-01-05T17:00:00Z").toZonedDateTimeISO(
    UTC,
  ),
  duration: durations.ONE_HOUR,
  hasSpecificTimes: false,
  timeIncrement: Temporal.Duration.from({ minutes: 30 }),
  eventTimezone: "America/Los_Angeles",
  slotGeneration: {
    startTimeLocal: Temporal.PlainTime.from("09:00"),
    endTimeLocal: Temporal.PlainTime.from("10:00"),
    timeIncrement: Temporal.Duration.from({ minutes: 30 }),
  },
  timedRecurrence: {
    kind: "weekly",
    selectedDays: [
      Temporal.PlainDate.from("2026-01-05"),
      Temporal.PlainDate.from("2026-01-07"),
    ],
    selectedDaysOfWeek: [1, 3],
    startOnMonday: true,
  },
  times: [
    Temporal.Instant.from("2026-01-05T17:00:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-01-05T17:30:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-01-07T17:00:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-01-07T17:30:00Z").toZonedDateTimeISO(UTC),
  ],
  activeSlots: [
    Temporal.Instant.from("2026-01-05T17:00:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-01-05T17:30:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-01-07T17:00:00Z").toZonedDateTimeISO(UTC),
    Temporal.Instant.from("2026-01-07T17:30:00Z").toZonedDateTimeISO(UTC),
  ],
})

describe("specificTimesEditDraft", () => {
  it("creates a full-day enabled domain when entering specific-times", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [Temporal.PlainDate.from("2026-05-30")],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("17:00"),
      timezoneValue: UTC,
      timeIncrementMinutes: 30,
    })

    const draft = buildSpecificTimesCreateDraft({
      schedule,
      timeIncrementMinutes: 30,
    })

    expect(draft.slotGeneration).toEqual({
      startTimeLocal: Temporal.PlainTime.from("00:00"),
      endTimeLocal: Temporal.PlainTime.from("00:00"),
      timeIncrement: Temporal.Duration.from({ minutes: 30 }),
    })
    expect(draft.enabledSlots).toHaveLength(48)
    expect(draft.enabledSlots?.[0]?.toInstant().toString()).toBe(
      "2026-05-30T00:00:00Z",
    )
    expect(draft.enabledSlots?.at(-1)?.toInstant().toString()).toBe(
      "2026-05-30T23:30:00Z",
    )
  })

  it("marks specific-time edits for reset when the slot window changes", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("17:00"),
      timezoneValue: UTC,
    })

    expect(
      buildSpecificTimesEditDraft({
        event: buildEvent(),
        schedule,
        timeIncrementMinutes: 15,
        specificTimesEnabled: true,
      }),
    ).toMatchObject({
      dates: schedule.normalizedSelectedDays,
      timeIncrementMinutes: 15,
      resetExistingTimes: true,
    })
  })

  it("preserves the prior subset on unchanged dates and keeps added dates enabled-only", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
        Temporal.PlainDate.from("2026-06-01"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    const draft = buildSpecificTimesEditDraft({
      event: buildEvent(),
      schedule,
      timeIncrementMinutes: 15,
      specificTimesEnabled: true,
    })

    expect(draft?.resetExistingTimes).toBe(false)
    expect(draft?.enabledSlots).toHaveLength(3 * 96)
    expect(draft?.activeSlots).toHaveLength(2)
    expect(draft?.activeSlots?.map((slot) => slot.toString())).toEqual([
      "2026-05-31T09:00:00+00:00[UTC]",
      "2026-05-31T09:15:00+00:00[UTC]",
    ])
    expect(
      draft?.activeSlots?.some(
        (slot) => slot.toPlainDate().toString() === "2026-06-01",
      ),
    ).toBe(false)
  })

  it("starts with an empty active subset when specific-times edit state resets", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("17:00"),
      timezoneValue: UTC,
      timeIncrementMinutes: 30,
    })

    const draft = buildSpecificTimesEditDraft({
      event: buildEvent(),
      schedule,
      timeIncrementMinutes: 30,
      specificTimesEnabled: true,
    })

    expect(draft?.resetExistingTimes).toBe(true)
    expect(draft?.enabledSlots).toHaveLength(2 * 48)
    expect(draft?.enabledSlots?.[0]?.toString()).toBe(
      "2026-05-30T00:00:00+00:00[UTC]",
    )
    expect(draft?.enabledSlots?.at(-1)?.toString()).toBe(
      "2026-05-31T23:30:00+00:00[UTC]",
    )
    expect(draft?.activeSlots).toEqual([])
  })

  it("filters the active subset down to remaining picked dates when membership shrinks", () => {
    const event = buildEvent()
    event.dates = [
      Temporal.PlainDate.from("2026-05-28"),
      Temporal.PlainDate.from("2026-05-29"),
      Temporal.PlainDate.from("2026-05-30"),
    ]
    event.timedRecurrence = {
      kind: "specific_dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-28"),
        Temporal.PlainDate.from("2026-05-29"),
        Temporal.PlainDate.from("2026-05-30"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
    }
    event.activeSlots = [
      Temporal.Instant.from("2026-05-29T09:00:00Z").toZonedDateTimeISO(UTC),
      Temporal.Instant.from("2026-05-29T09:15:00Z").toZonedDateTimeISO(UTC),
    ]

    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-29"),
        Temporal.PlainDate.from("2026-05-30"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    expect(
      buildSpecificTimesEditDraft({
        event,
        schedule,
        timeIncrementMinutes: 15,
        specificTimesEnabled: true,
      }),
    ).toMatchObject({
      resetExistingTimes: false,
      activeSlots: [
        Temporal.Instant.from("2026-05-29T09:00:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-05-29T09:15:00Z").toZonedDateTimeISO(UTC),
      ],
    })
  })

  it("preserves saved specific times when membership dates and increment stay unchanged", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    expect(
      buildSpecificTimesEditDraft({
        event: buildEvent(),
        schedule,
        timeIncrementMinutes: 15,
        specificTimesEnabled: true,
      }),
    ).toMatchObject({
      resetExistingTimes: false,
      activeSlots: [
        Temporal.Instant.from("2026-05-31T09:00:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-05-31T09:15:00Z").toZonedDateTimeISO(UTC),
      ],
    })
  })

  it("keeps the saved active subset when canonical fields still match stale legacy dates", () => {
    const event = buildEvent()
    event.dates = [
      Temporal.PlainDate.from("2026-05-01"),
      Temporal.PlainDate.from("2026-05-02"),
    ]

    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    expect(
      buildSpecificTimesEditDraft({
        event,
        schedule,
        timeIncrementMinutes: 15,
        specificTimesEnabled: true,
      }),
    ).toMatchObject({
      resetExistingTimes: false,
      activeSlots: [
        Temporal.Instant.from("2026-05-31T09:00:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-05-31T09:15:00Z").toZonedDateTimeISO(UTC),
      ],
    })
  })

  it("discards all actives when a cross-midnight timezone change leaves none in the rebuilt domain", () => {
    const event: Event = {
      ...buildEvent(),
      eventTimezone: "America/Los_Angeles",
      dates: [Temporal.PlainDate.from("2026-01-04")],
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: [Temporal.PlainDate.from("2026-01-04")],
        selectedDaysOfWeek: [],
        startOnMonday: true,
      },
      activeSlots: [
        Temporal.Instant.from("2026-01-05T07:30:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-01-05T08:00:00Z").toZonedDateTimeISO(UTC),
      ],
      times: [
        Temporal.Instant.from("2026-01-05T07:30:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-01-05T08:00:00Z").toZonedDateTimeISO(UTC),
      ],
      slotGeneration: {
        startTimeLocal: Temporal.PlainTime.from("23:30"),
        endTimeLocal: Temporal.PlainTime.from("01:30"),
        timeIncrement: Temporal.Duration.from({ minutes: 30 }),
      },
    }

    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [Temporal.PlainDate.from("2026-01-04")],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("23:30"),
      endTime: Temporal.PlainTime.from("01:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 30,
    })

    const draft = buildSpecificTimesEditDraft({
      event,
      schedule,
      timeIncrementMinutes: 30,
      specificTimesEnabled: true,
    })

    expect(draft?.resetExistingTimes).toBe(false)
    expect(draft?.dates?.map((day) => day.toString())).toEqual(["2026-01-04"])
    expect(
      draft?.timedRecurrence?.selectedDays?.map((day) => day.toString()),
    ).toEqual(["2026-01-04"])
    expect(
      draft?.activeSlots?.map((slot) => slot.toInstant().toString()),
    ).toEqual([])
  })

  it("keeps only in-domain actives when a cross-midnight timezone change discards the rest", () => {
    const event: Event = {
      ...buildEvent(),
      eventTimezone: "America/Los_Angeles",
      dates: [Temporal.PlainDate.from("2026-01-04")],
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: [Temporal.PlainDate.from("2026-01-04")],
        selectedDaysOfWeek: [],
        startOnMonday: true,
      },
      activeSlots: [
        Temporal.Instant.from("2026-01-04T18:00:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-01-05T07:30:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-01-05T08:00:00Z").toZonedDateTimeISO(UTC),
      ],
      times: [
        Temporal.Instant.from("2026-01-04T18:00:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-01-05T07:30:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-01-05T08:00:00Z").toZonedDateTimeISO(UTC),
      ],
      slotGeneration: {
        startTimeLocal: Temporal.PlainTime.from("23:30"),
        endTimeLocal: Temporal.PlainTime.from("01:30"),
        timeIncrement: Temporal.Duration.from({ minutes: 30 }),
      },
    }

    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [Temporal.PlainDate.from("2026-01-04")],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("23:30"),
      endTime: Temporal.PlainTime.from("01:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 30,
    })

    const draft = buildSpecificTimesEditDraft({
      event,
      schedule,
      timeIncrementMinutes: 30,
      specificTimesEnabled: true,
    })

    expect(draft?.resetExistingTimes).toBe(false)
    expect(draft?.dates?.map((day) => day.toString())).toEqual(["2026-01-04"])
    expect(
      draft?.timedRecurrence?.selectedDays?.map((day) => day.toString()),
    ).toEqual(["2026-01-04"])
    expect(
      draft?.activeSlots?.map((slot) => slot.toInstant().toString()),
    ).toEqual(["2026-01-04T18:00:00Z"])
  })

  it("keeps picked dates stable and filters the active subset when only the canonical timezone changes", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: "America/Los_Angeles",
      timeIncrementMinutes: 15,
    })

    const draft = buildSpecificTimesEditDraft({
      event: buildEvent(),
      schedule,
      timeIncrementMinutes: 15,
      specificTimesEnabled: true,
    })

    expect(draft?.resetExistingTimes).toBe(false)
    expect(draft?.enabledSlots).toHaveLength(2 * 96)
    expect(draft?.enabledSlots?.[0]?.toInstant().toString()).toBe(
      "2026-05-30T07:00:00Z",
    )
    expect(draft?.activeSlots?.map((slot) => slot.toString())).toEqual([
      "2026-05-31T09:00:00+00:00[UTC]",
      "2026-05-31T09:15:00+00:00[UTC]",
    ])
  })

  it("restores active slots to the full civil-day enabled domain when specific-times editing is disabled", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    const draft = buildSpecificTimesEditDraft({
      event: buildEvent(),
      schedule,
      timeIncrementMinutes: 15,
      specificTimesEnabled: false,
    })
    expect(draft?.resetExistingTimes).toBe(false)
    expect(draft?.enabledSlots).toHaveLength(2 * 96)
    expect(draft?.activeSlots).toHaveLength(2 * 96)
    expect(draft?.enabledSlots?.[0]?.toString()).toBe(
      "2026-05-30T00:00:00+00:00[UTC]",
    )
    expect(draft?.enabledSlots?.at(-1)?.toString()).toBe(
      "2026-05-31T23:45:00+00:00[UTC]",
    )
  })

  it("rewrites non-specific timed edits to the schedule canonical slots instead of preserving stale out-of-window slots", () => {
    const event = buildEvent()
    event.hasSpecificTimes = false
    event.duration = Temporal.Duration.from({ hours: 8 })
    event.slotGeneration = {
      startTimeLocal: Temporal.PlainTime.from("09:00"),
      endTimeLocal: Temporal.PlainTime.from("17:00"),
      timeIncrement: durations.FIFTEEN_MINUTES,
    }
    event.activeSlots = [
      Temporal.Instant.from("2026-05-30T08:00:00Z").toZonedDateTimeISO(UTC),
      Temporal.Instant.from("2026-05-30T08:15:00Z").toZonedDateTimeISO(UTC),
      Temporal.Instant.from("2026-05-30T09:00:00Z").toZonedDateTimeISO(UTC),
      Temporal.Instant.from("2026-05-31T08:00:00Z").toZonedDateTimeISO(UTC),
      Temporal.Instant.from("2026-05-31T16:45:00Z").toZonedDateTimeISO(UTC),
    ]
    event.times = [
      Temporal.Instant.from("2026-05-30T09:00:00Z").toZonedDateTimeISO(UTC),
      Temporal.Instant.from("2026-05-31T09:00:00Z").toZonedDateTimeISO(UTC),
    ]

    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("17:00"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    const draft = buildSpecificTimesEditDraft({
      event,
      schedule,
      timeIncrementMinutes: 15,
      specificTimesEnabled: false,
    })

    expect(
      draft?.enabledSlots?.map((slot) => slot.toInstant().toString()),
    ).toEqual(schedule.enabledSlots.map((slot) => slot.toInstant().toString()))
    expect(
      draft?.activeSlots?.map((slot) => slot.toInstant().toString()),
    ).toEqual(schedule.enabledSlots.map((slot) => slot.toInstant().toString()))
    expect(
      draft?.enabledSlots?.some((slot) =>
        slot.toInstant().toString().includes("T08:"),
      ),
    ).toBe(false)
    expect(
      draft?.activeSlots?.some((slot) =>
        slot.toInstant().toString().includes("T08:"),
      ),
    ).toBe(false)
  })

  it("applies a reset draft by clearing stale times and updating local event seeds", () => {
    const event = buildEvent()
    const nextEvent = applySpecificTimesEditDraft({
      event,
      draft: {
        dates: [
          Temporal.PlainDate.from("2026-05-28"),
          Temporal.PlainDate.from("2026-05-29"),
        ],
        timeSeed: Temporal.Instant.from(
          "2026-05-28T09:00:00Z",
        ).toZonedDateTimeISO(UTC),
        duration: Temporal.Duration.from({ hours: 8 }),
        timeIncrementMinutes: 30,
        resetExistingTimes: true,
      },
    })

    expect(nextEvent.dates?.map((date) => date.toString())).toEqual([
      "2026-05-28",
      "2026-05-29",
    ])
    expect(nextEvent.timeSeed?.toString()).toBe(
      "2026-05-28T09:00:00+00:00[UTC]",
    )
    expect(nextEvent.timeIncrement?.total("minutes")).toBe(30)
    expect(nextEvent.times).toEqual([])
    expect(processEventMock).toHaveBeenCalledWith(nextEvent)
  })

  it("applies weekly timed drafts without relabeling them as specific_dates", () => {
    const event = buildWeeklyEvent()
    const nextEvent = applySpecificTimesEditDraft({
      event,
      draft: {
        dates: [...(event.dates ?? [])],
        timeSeed: event.timeSeed,
        duration: event.duration,
        activeSlots: [...(event.activeSlots ?? [])],
        eventTimezone: event.eventTimezone,
        timedRecurrence: event.timedRecurrence,
        slotGeneration: event.slotGeneration,
        timeIncrementMinutes: 30,
        resetExistingTimes: false,
      },
    })

    expect(nextEvent.type).toBe(eventTypes.DOW)
    expect(nextEvent.timedRecurrence?.kind).toBe("weekly")
  })

  it("applies specific-date timed drafts with the canonical specific_dates type", () => {
    const nextEvent = applySpecificTimesEditDraft({
      event: buildWeeklyEvent(),
      draft: {
        dates: [
          Temporal.PlainDate.from("2026-05-30"),
          Temporal.PlainDate.from("2026-05-31"),
        ],
        timeSeed: Temporal.Instant.from(
          "2026-05-30T09:00:00Z",
        ).toZonedDateTimeISO(UTC),
        duration: durations.ONE_HOUR,
        enabledSlots: [
          Temporal.Instant.from("2026-05-30T09:00:00Z").toZonedDateTimeISO(UTC),
          Temporal.Instant.from("2026-05-30T09:15:00Z").toZonedDateTimeISO(UTC),
          Temporal.Instant.from("2026-05-31T09:00:00Z").toZonedDateTimeISO(UTC),
          Temporal.Instant.from("2026-05-31T09:15:00Z").toZonedDateTimeISO(UTC),
        ],
        activeSlots: [
          Temporal.Instant.from("2026-05-31T09:00:00Z").toZonedDateTimeISO(UTC),
          Temporal.Instant.from("2026-05-31T09:15:00Z").toZonedDateTimeISO(UTC),
        ],
        eventTimezone: UTC,
        timedRecurrence: {
          kind: "specific_dates",
          selectedDays: [
            Temporal.PlainDate.from("2026-05-30"),
            Temporal.PlainDate.from("2026-05-31"),
          ],
          selectedDaysOfWeek: [],
          startOnMonday: true,
        },
        slotGeneration: {
          startTimeLocal: Temporal.PlainTime.from("09:00"),
          endTimeLocal: Temporal.PlainTime.from("09:30"),
          timeIncrement: durations.FIFTEEN_MINUTES,
        },
        timeIncrementMinutes: 15,
        resetExistingTimes: false,
      },
    })

    expect(nextEvent.type).toBe(eventTypes.SPECIFIC_DATES)
    expect(nextEvent.timedRecurrence?.kind).toBe("specific_dates")
  })

  it("keeps type and timed recurrence aligned after draft application", () => {
    const weeklyEvent = applySpecificTimesEditDraft({
      event: buildWeeklyEvent(),
      draft: {
        dates: [...(buildWeeklyEvent().dates ?? [])],
        timeSeed: buildWeeklyEvent().timeSeed,
        duration: buildWeeklyEvent().duration,
        activeSlots: [...(buildWeeklyEvent().activeSlots ?? [])],
        eventTimezone: buildWeeklyEvent().eventTimezone,
        timedRecurrence: buildWeeklyEvent().timedRecurrence,
        slotGeneration: buildWeeklyEvent().slotGeneration,
        timeIncrementMinutes: 30,
        resetExistingTimes: false,
      },
    })
    const specificDatesEvent = applySpecificTimesEditDraft({
      event: buildEvent(),
      draft: {
        dates: [...(buildEvent().dates ?? [])],
        timeSeed: buildEvent().timeSeed,
        duration: buildEvent().duration,
        activeSlots: [...(buildEvent().activeSlots ?? [])],
        eventTimezone: buildEvent().eventTimezone,
        timedRecurrence: buildEvent().timedRecurrence,
        slotGeneration: buildEvent().slotGeneration,
        timeIncrementMinutes: 15,
        resetExistingTimes: false,
      },
    })

    expect(weeklyEvent.type).toBe(eventTypes.DOW)
    expect(weeklyEvent.timedRecurrence?.kind).toBe("weekly")
    expect(specificDatesEvent.type).toBe(eventTypes.SPECIFIC_DATES)
    expect(specificDatesEvent.timedRecurrence?.kind).toBe("specific_dates")
  })

  it("keeps only the saved active subset when rebuilding a specific-times draft", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    const draft = buildSpecificTimesEditDraft({
      event: buildEvent(),
      schedule,
      timeIncrementMinutes: 15,
      specificTimesEnabled: true,
    })

    expect(draft?.enabledSlots).toHaveLength(2 * 96)
    expect(draft?.enabledSlots?.[0]?.toString()).toBe(
      "2026-05-30T00:00:00+00:00[UTC]",
    )
    expect(draft?.enabledSlots?.at(-1)?.toString()).toBe(
      "2026-05-31T23:45:00+00:00[UTC]",
    )
    expect(draft?.activeSlots?.map((slot) => slot.toString())).toEqual([
      "2026-05-31T09:00:00+00:00[UTC]",
      "2026-05-31T09:15:00+00:00[UTC]",
    ])
    expect(draft?.slotGeneration).toEqual({
      startTimeLocal: Temporal.PlainTime.from("09:00"),
      endTimeLocal: Temporal.PlainTime.from("09:30"),
      timeIncrement: Temporal.Duration.from({ minutes: 15 }),
    })
  })

  it("filters out enabledSlots for removed dates when dates changed", () => {
    const eventWithThreeDates: Event = {
      ...buildEvent(),
      dates: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
        Temporal.PlainDate.from("2026-06-01"),
      ],
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: [
          Temporal.PlainDate.from("2026-05-30"),
          Temporal.PlainDate.from("2026-05-31"),
          Temporal.PlainDate.from("2026-06-01"),
        ],
        selectedDaysOfWeek: [],
        startOnMonday: true,
      },
      times: [
        Temporal.Instant.from("2026-05-30T00:00:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-05-31T00:00:00Z").toZonedDateTimeISO(UTC),
        Temporal.Instant.from("2026-06-01T00:00:00Z").toZonedDateTimeISO(UTC),
      ],
    }

    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Specific dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
      selectedDaysOfWeek: [],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("09:30"),
      timezoneValue: UTC,
      timeIncrementMinutes: 15,
    })

    const draft = buildSpecificTimesEditDraft({
      event: eventWithThreeDates,
      schedule,
      timeIncrementMinutes: 15,
      specificTimesEnabled: true,
    })

    expect(draft?.dates?.map((d) => d.toString())).toEqual([
      "2026-05-30",
      "2026-05-31",
    ])
    expect(draft?.timedRecurrence).toMatchObject({
      kind: "specific_dates",
      selectedDays: [
        Temporal.PlainDate.from("2026-05-30"),
        Temporal.PlainDate.from("2026-05-31"),
      ],
    })
    expect(draft?.enabledSlots).toHaveLength(2 * 96)
    expect(
      draft?.enabledSlots?.every(
        (slot) => slot.toPlainDate().toString() !== "2026-06-01",
      ),
    ).toBe(true)
    // activeSlots keep the prior subset after removed-date filtering
    expect(draft?.activeSlots?.map((slot) => slot.toString())).toEqual([
      "2026-05-31T09:00:00+00:00[UTC]",
      "2026-05-31T09:15:00+00:00[UTC]",
    ])
  })

  it("preserves canonical weekly selectedDays on edit drafts even when the schedule builder emits an empty weekly list", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Days of the week",
      selectedDays: [],
      selectedDaysOfWeek: [1, 3],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("10:00"),
      timezoneValue: "America/Los_Angeles",
      timeIncrementMinutes: 30,
    })

    const draft = buildSpecificTimesEditDraft({
      event: buildWeeklyEvent(),
      schedule,
      timeIncrementMinutes: 30,
      specificTimesEnabled: false,
    })

    expect(draft?.timedRecurrence).toMatchObject({
      kind: "weekly",
      selectedDays: [
        Temporal.PlainDate.from("2026-01-05"),
        Temporal.PlainDate.from("2026-01-07"),
      ],
      selectedDaysOfWeek: [1, 3],
      startOnMonday: true,
    })
  })

  it("preserves the event's active slots when rebuilding a weekly draft without specific times", () => {
    const schedule = buildEventEditorSchedule({
      daysOnly: false,
      daysOnlyType: "specific_dates",
      selectedDateOption: "Days of the week",
      selectedDays: [],
      selectedDaysOfWeek: [1, 3],
      startOnMonday: true,
      startTime: Temporal.PlainTime.from("09:00"),
      endTime: Temporal.PlainTime.from("10:00"),
      timezoneValue: "America/Los_Angeles",
      timeIncrementMinutes: 30,
      weeklyAnchorInstant: buildWeeklyEvent().activeSlots?.[0]?.withTimeZone(
        "America/Los_Angeles",
      ),
    })

    const draft = buildSpecificTimesEditDraft({
      event: buildWeeklyEvent(),
      schedule,
      timeIncrementMinutes: 30,
      specificTimesEnabled: false,
    })

    expect(
      draft?.activeSlots?.map((slot) => slot.toInstant().toString()),
    ).toEqual([
      "2026-01-05T17:00:00Z",
      "2026-01-05T17:30:00Z",
      "2026-01-07T17:00:00Z",
      "2026-01-07T17:30:00Z",
    ])
  })
})
