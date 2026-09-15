import { computed, ref } from "vue"
import { Temporal } from "temporal-polyfill"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type * as UtilsModule from "@/utils"
import type { ScheduleOverlapEvent } from "./types"
import { useSignUpForm } from "./useSignUpForm"

const { putMock, showErrorMock } = vi.hoisted(() => ({
  putMock: vi.fn(),
  showErrorMock: vi.fn(),
}))

vi.mock("@/utils", async () => {
  const actual = await vi.importActual<typeof UtilsModule>("@/utils")

  return {
    ...actual,
    put: putMock,
  }
})

vi.mock("@/stores/main", () => ({
  useMainStore: () => ({
    showError: showErrorMock,
    authUser: null,
  }),
}))

describe("useSignUpForm", () => {
  beforeEach(() => {
    putMock.mockReset()
    showErrorMock.mockReset()
    putMock.mockResolvedValue(undefined)
  })

  it("preserves canonical timed fields when saving edited sign-up blocks", async () => {
    const event = ref<ScheduleOverlapEvent>({
      _id: "evt-1",
      name: "Timed sign up",
      type: "specific_dates",
      duration: Temporal.Duration.from({ hours: 1 }),
      dates: [Temporal.PlainDate.from("2026-05-28")],
      timeSeed: Temporal.ZonedDateTime.from("2026-05-28T09:00:00+00:00[UTC]"),
      activeSlots: [
        Temporal.ZonedDateTime.from("2026-05-28T09:15:00+00:00[UTC]"),
        Temporal.ZonedDateTime.from("2026-05-28T09:30:00+00:00[UTC]"),
      ],
      eventTimezone: "UTC",
      slotGeneration: {
        startTimeLocal: Temporal.PlainTime.from("09:00"),
        endTimeLocal: Temporal.PlainTime.from("10:00"),
        timeIncrement: Temporal.Duration.from({ minutes: 15 }),
      },
      timedRecurrence: {
        kind: "specific_dates" as const,
        selectedDays: [Temporal.PlainDate.from("2026-05-28")],
        selectedDaysOfWeek: [],
        startOnMonday: true,
      },
      signUpBlocks: [],
    })

    const form = useSignUpForm({
      event,
      isSignUp: computed(() => true),
      days: computed(() => []),
      isOwner: computed(() => true),
      dragStart: ref(null),
    })

    form.signUpBlocksByDay.value = [
      [
        {
          _id: "block-1",
          name: "Slot 1",
          capacity: 2,
          startDate: Temporal.ZonedDateTime.from(
            "2026-05-28T09:15:00+00:00[UTC]",
          ),
          endDate: Temporal.ZonedDateTime.from(
            "2026-05-28T09:45:00+00:00[UTC]",
          ),
          hoursOffset: Temporal.Duration.from({ minutes: 15 }),
          hoursLength: Temporal.Duration.from({ minutes: 30 }),
        },
      ],
    ]
    form.signUpBlocksToAddByDay.value = [[]]

    await expect(form.submitNewSignUpBlocks()).resolves.toBe(true)

    expect(putMock).toHaveBeenCalledWith("/events/evt-1", {
      activeSlots: ["2026-05-28T09:15:00Z", "2026-05-28T09:30:00Z"],
      eventTimezone: "UTC",
      slotGeneration: {
        startTimeLocal: "09:00:00",
        endTimeLocal: "10:00:00",
        timeIncrementMinutes: 15,
      },
      timedRecurrence: {
        kind: "specific_dates",
        selectedDays: ["2026-05-28"],
        selectedDaysOfWeek: [],
        startOnMonday: true,
      },
      name: "Timed sign up",
      notificationsEnabled: undefined,
      blindAvailabilityEnabled: undefined,
      daysOnly: undefined,
      type: "specific_dates",
      sendEmailAfterXResponses: undefined,
      collectEmails: undefined,
      creatorPosthogId: undefined,
      description: undefined,
      signUpBlocks: [
        {
          _id: "block-1",
          name: "Slot 1",
          capacity: 2,
          startDate: Temporal.Instant.from("2026-05-28T09:15:00Z")
            .epochMilliseconds,
          endDate: Temporal.Instant.from("2026-05-28T09:45:00Z")
            .epochMilliseconds,
        },
      ],
      remindees: undefined,
      attendees: undefined,
    })
  })

  it("converts bucketed sign-up blocks to Duration offsets when resetting", () => {
    const event = ref<ScheduleOverlapEvent>({
      _id: "evt-2",
      name: "Timed sign up",
      type: "specific_dates",
      duration: Temporal.Duration.from({ hours: 8 }),
      dates: [Temporal.PlainDate.from("2026-05-28")],
      timeSeed: Temporal.ZonedDateTime.from("2026-05-28T09:00:00+00:00[UTC]"),
      activeSlots: [
        Temporal.ZonedDateTime.from("2026-05-28T09:00:00+00:00[UTC]"),
      ],
      eventTimezone: "UTC",
      slotGeneration: {
        startTimeLocal: Temporal.PlainTime.from("09:00"),
        endTimeLocal: Temporal.PlainTime.from("17:00"),
        timeIncrement: Temporal.Duration.from({ minutes: 60 }),
      },
      timedRecurrence: {
        kind: "specific_dates" as const,
        selectedDays: [Temporal.PlainDate.from("2026-05-28")],
        selectedDaysOfWeek: [],
        startOnMonday: true,
      },
      signUpBlocks: [
        {
          _id: "block-2",
          name: "Morning Slot",
          capacity: 1,
          startDate: Temporal.ZonedDateTime.from(
            "2026-05-28T10:30:00+00:00[UTC]",
          ),
          endDate: Temporal.ZonedDateTime.from(
            "2026-05-28T11:30:00+00:00[UTC]",
          ),
          hoursOffset: Temporal.Duration.from({ minutes: 0 }),
          hoursLength: Temporal.Duration.from({ minutes: 0 }),
        },
      ],
    })

    const form = useSignUpForm({
      event,
      isSignUp: computed(() => true),
      days: computed(() => []),
      isOwner: computed(() => true),
      dragStart: ref(null),
    })

    form.resetSignUpForm()

    const [dayBlocks] = form.signUpBlocksByDay.value
    expect(dayBlocks).toHaveLength(1)
    expect(dayBlocks[0].hoursOffset).toEqual(
      Temporal.Duration.from({ minutes: 90 }),
    )
    expect(dayBlocks[0].hoursLength).toEqual(
      Temporal.Duration.from({ minutes: 60 }),
    )
    expect(typeof dayBlocks[0].hoursOffset).not.toBe("number")
  })

  it("mints unique non-canonical local ids for new sign-up blocks", () => {
    const event = ref<ScheduleOverlapEvent>({
      _id: "evt-3",
      name: "Timed sign up",
      type: "specific_dates",
      signUpBlocks: [],
    })

    const form = useSignUpForm({
      event,
      isSignUp: computed(() => true),
      days: computed(() => [
        {
          dayText: "Thu",
          dateString: "2026-05-28",
          dateObject: Temporal.ZonedDateTime.from(
            "2026-05-28T00:00:00+00:00[UTC]",
          ),
        },
      ]),
      isOwner: computed(() => true),
      dragStart: ref(null),
    })

    const first = form.createSignUpBlock(
      0,
      Temporal.Duration.from({ minutes: 0 }),
      Temporal.Duration.from({ minutes: 30 }),
    )
    const second = form.createSignUpBlock(
      0,
      Temporal.Duration.from({ minutes: 30 }),
      Temporal.Duration.from({ minutes: 30 }),
    )

    expect(first._id).not.toBe(second._id)
    expect(first._id).not.toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    )
    expect(first.startDate).toBeInstanceOf(Temporal.ZonedDateTime)
    expect(first.endDate).toBeInstanceOf(Temporal.ZonedDateTime)
  })
})
