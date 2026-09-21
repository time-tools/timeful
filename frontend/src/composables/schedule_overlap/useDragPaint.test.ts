// @vitest-environment happy-dom

import { computed, ref, shallowRef } from "vue"
import { describe, expect, it, vi } from "vitest"
import { Temporal } from "temporal-polyfill"
import { availabilityTypes, eventTypes, UTC } from "@/constants"
import { ZdtMap, ZdtSet } from "@/utils"
import {
  states,
  type ScheduledEvent,
} from "@/composables/schedule_overlap/types"
import { useDragPaint } from "./useDragPaint"

const zdt = (iso: string) => Temporal.Instant.from(iso).toZonedDateTimeISO(UTC)

function createScheduleEventDragPaint(activeRows = [0]) {
  const event = ref({
    _id: "evt-1",
    ownerId: "owner-1",
    name: "Schedule event",
    type: eventTypes.SPECIFIC_DATES,
    dates: [Temporal.PlainDate.from("2026-01-01")],
    timeSeed: zdt("2026-01-01T09:00:00Z"),
    daysOnly: false,
    hasSpecificTimes: false,
  })
  const state = ref(states.SCHEDULE_EVENT)
  const dragging = ref(false)
  const dragStart = ref<{ row: number; col: number } | null>(null)
  const dragCur = ref<{ row: number; col: number } | null>(null)
  const curScheduledEvent = ref<ScheduledEvent | null>(null)

  return {
    dragPaint: useDragPaint({
      event,
      state,
      isSignUp: computed(() => false),
      weekOffset: ref(0),
      dragging,
      dragStart,
      dragCur,
      curTimeslot: ref({ row: -1, col: -1 }),
      splitTimes: computed(() => [
        Array.from({ length: 4 }, (_, row) => ({
          hoursOffset: Temporal.Duration.from({ hours: 9 + row }),
        })),
        [],
      ]),
      times: computed(() =>
        Array.from({ length: 4 }, (_, row) => ({
          hoursOffset: Temporal.Duration.from({ hours: 9 + row }),
        })),
      ),
      days: computed(() => [{ dateObject: zdt("2026-01-01T09:00:00Z") }]),
      monthDays: computed(() => []),
      monthDayIncluded: computed(() => new ZdtMap<boolean>()),
      columnOffsets: computed(() => [10]),
      timeslot: ref({ width: 10, height: 10 }),
      availability: shallowRef(new ZdtSet()),
      ifNeeded: shallowRef(new ZdtSet()),
      tempTimes: shallowRef(new ZdtSet()),
      availabilityType: ref(availabilityTypes.AVAILABLE),
      signUpBlocksByDay: ref([[]]),
      signUpBlocksToAddByDay: ref([[]]),
      manualAvailability: shallowRef(new ZdtMap<ZdtSet>()),
      setScheduledEventFromRowCol: (scheduledEvent) => {
        curScheduledEvent.value = scheduledEvent
      },
      maxSignUpBlockRowSize: computed(() => null),
      allowDrag: computed(() => true),
      getDateFromRowCol: (row, col) =>
        activeRows.includes(row)
          ? zdt(`2026-01-01T${String(9 + row).padStart(2, "0")}:00:00Z`).add({
              days: col,
            })
          : null,
      getAvailabilityForColumn: () => new ZdtSet(),
      createSignUpBlock: () => {
        throw new Error("not used in schedule-event tests")
      },
    }),
    curScheduledEvent,
    dragging,
  }
}

function createAvailabilityDragPaint() {
  const event = ref({
    _id: "evt-2",
    ownerId: "owner-1",
    name: "Edit availability",
    type: eventTypes.SPECIFIC_DATES,
    dates: [Temporal.PlainDate.from("2026-01-01")],
    timeSeed: zdt("2026-01-01T09:00:00Z"),
    daysOnly: false,
    hasSpecificTimes: false,
  })
  const state = ref(states.EDIT_AVAILABILITY)
  const dragging = ref(false)
  const dragStart = ref<{ row: number; col: number } | null>(null)
  const dragCur = ref<{ row: number; col: number } | null>(null)
  const curTimeslot = ref({ row: -1, col: -1 })

  return {
    dragPaint: useDragPaint({
      event,
      state,
      isSignUp: computed(() => false),
      weekOffset: ref(0),
      dragging,
      dragStart,
      dragCur,
      curTimeslot,
      splitTimes: computed(() => [
        [{ hoursOffset: Temporal.Duration.from({ hours: 9 }) }],
        [],
      ]),
      times: computed(() => [
        { hoursOffset: Temporal.Duration.from({ hours: 9 }) },
        { hoursOffset: Temporal.Duration.from({ hours: 10 }) },
        { hoursOffset: Temporal.Duration.from({ hours: 11 }) },
        { hoursOffset: Temporal.Duration.from({ hours: 12 }) },
      ]),
      days: computed(() => [{ dateObject: zdt("2026-01-01T09:00:00Z") }]),
      monthDays: computed(() => []),
      monthDayIncluded: computed(() => new ZdtMap<boolean>()),
      columnOffsets: computed(() => [10]),
      timeslot: ref({ width: 10, height: 10 }),
      availability: shallowRef(new ZdtSet()),
      ifNeeded: shallowRef(new ZdtSet()),
      tempTimes: shallowRef(new ZdtSet()),
      availabilityType: ref(availabilityTypes.AVAILABLE),
      signUpBlocksByDay: ref([[]]),
      signUpBlocksToAddByDay: ref([[]]),
      manualAvailability: shallowRef(new ZdtMap<ZdtSet>()),
      setScheduledEventFromRowCol: vi.fn(),
      maxSignUpBlockRowSize: computed(() => null),
      allowDrag: computed(() => true),
      getDateFromRowCol: (row, col) =>
        zdt(`2026-01-01T${String(9 + row).padStart(2, "0")}:00:00Z`).add({
          days: col,
        }),
      getAvailabilityForColumn: () => new ZdtSet(),
      createSignUpBlock: () => {
        throw new Error("not used in edit-availability tests")
      },
    }),
    curTimeslot,
  }
}

function createSpecificTimesDragPaint() {
  const event = ref({
    _id: "evt-3",
    ownerId: "owner-1",
    name: "Specific times",
    type: eventTypes.SPECIFIC_DATES,
    dates: [
      Temporal.PlainDate.from("2026-05-29"),
      Temporal.PlainDate.from("2026-05-30"),
    ],
    timeSeed: zdt("2026-05-29T00:00:00Z"),
    daysOnly: false,
    hasSpecificTimes: true,
  })
  const state = ref(states.SET_SPECIFIC_TIMES)
  const dragging = ref(false)
  const dragStart = ref<{ row: number; col: number } | null>(null)
  const dragCur = ref<{ row: number; col: number } | null>(null)
  const tempTimes = shallowRef(new ZdtSet())
  const curTimeslot = ref({ row: -1, col: -1 })

  return {
    dragPaint: useDragPaint({
      event,
      state,
      isSignUp: computed(() => false),
      weekOffset: ref(0),
      dragging,
      dragStart,
      dragCur,
      curTimeslot,
      splitTimes: computed(() => [
        Array.from({ length: 16 }, (_, row) => ({
          hoursOffset: Temporal.Duration.from({ minutes: row * 15 }),
          text: String(row),
          absoluteMinutes: row * 15,
          displayedMinutes: row * 15,
        })),
        [],
      ]),
      times: computed(() =>
        Array.from({ length: 16 }, (_, row) => ({
          hoursOffset: Temporal.Duration.from({ minutes: row * 15 }),
          text: String(row),
          absoluteMinutes: row * 15,
          displayedMinutes: row * 15,
        })),
      ),
      days: computed(() => [
        { dateObject: zdt("2026-05-29T00:00:00Z") },
        { dateObject: zdt("2026-05-30T00:00:00Z") },
      ]),
      monthDays: computed(() => []),
      monthDayIncluded: computed(() => new ZdtMap<boolean>()),
      columnOffsets: computed(() => [10, 20, 30]),
      timeslot: ref({ width: 10, height: 10 }),
      availability: shallowRef(new ZdtSet()),
      ifNeeded: shallowRef(new ZdtSet()),
      tempTimes,
      availabilityType: ref(availabilityTypes.AVAILABLE),
      signUpBlocksByDay: ref([[], []]),
      signUpBlocksToAddByDay: ref([[], []]),
      manualAvailability: shallowRef(new ZdtMap<ZdtSet>()),
      setScheduledEventFromRowCol: vi.fn(),
      maxSignUpBlockRowSize: computed(() => null),
      allowDrag: computed(() => true),
      getDateFromRowCol: (row, col) =>
        zdt(`2026-05-${String(29 + col).padStart(2, "0")}T00:00:00Z`).add({
          minutes: row * 15,
        }),
      getAvailabilityForColumn: () => new ZdtSet(),
      createSignUpBlock: () => {
        throw new Error("not used in specific-times tests")
      },
    }),
    dragStart,
    dragCur,
    tempTimes,
    curTimeslot,
  }
}

function createPointerEvent(
  type: string,
  target: HTMLElement,
  pointerId: number,
  y: number,
) {
  const event = new PointerEvent(type, {
    clientX: 5,
    clientY: y,
    pointerId,
  })
  Object.defineProperty(event, "currentTarget", {
    configurable: true,
    value: target,
  })
  return event
}

describe("useDragPaint pointer capture", () => {
  it("captures the active pointer for schedule-event drags and releases it on completion", () => {
    const { dragPaint, curScheduledEvent, dragging } =
      createScheduleEventDragPaint()
    const target = document.createElement("div")

    Object.defineProperty(target, "getBoundingClientRect", {
      configurable: true,
      value: () => ({
        left: 0,
        top: 0,
        width: 10,
        height: 10,
      }),
    })

    const setPointerCapture = vi.fn()
    const releasePointerCapture = vi.fn()
    const hasPointerCapture = vi.fn(() => true)
    target.setPointerCapture = setPointerCapture
    target.releasePointerCapture = releasePointerCapture
    target.hasPointerCapture = hasPointerCapture

    dragPaint.startDrag(createPointerEvent("pointerdown", target, 7, 5))
    dragPaint.moveDrag(createPointerEvent("pointermove", target, 7, 25))
    dragPaint.endDrag(createPointerEvent("pointerup", target, 7, 25))

    expect(setPointerCapture).toHaveBeenCalledWith(7)
    expect(releasePointerCapture).toHaveBeenCalledWith(7)
    expect(dragging.value).toBe(false)
    expect(curScheduledEvent.value).toEqual({
      col: 0,
      row: 0,
      numRows: 1,
    })
  })

  it("ignores pointer movement from a different pointer id during an active drag", () => {
    const { dragPaint, curScheduledEvent } = createScheduleEventDragPaint()
    const target = document.createElement("div")

    Object.defineProperty(target, "getBoundingClientRect", {
      configurable: true,
      value: () => ({
        left: 0,
        top: 0,
        width: 10,
        height: 10,
      }),
    })

    target.setPointerCapture = vi.fn()
    target.releasePointerCapture = vi.fn()
    target.hasPointerCapture = vi.fn(() => true)

    dragPaint.startDrag(createPointerEvent("pointerdown", target, 7, 5))
    dragPaint.moveDrag(createPointerEvent("pointermove", target, 11, 35))
    dragPaint.endDrag(createPointerEvent("pointerup", target, 7, 35))

    expect(curScheduledEvent.value).toEqual({
      col: 0,
      row: 0,
      numRows: 1,
    })
  })

  it("keeps schedule-event drags within active specific-time slots", () => {
    const { dragPaint, curScheduledEvent } = createScheduleEventDragPaint([
      0, 1,
    ])
    const target = document.createElement("div")

    Object.defineProperty(target, "getBoundingClientRect", {
      configurable: true,
      value: () => ({ left: 0, top: 0, width: 10, height: 10 }),
    })
    target.setPointerCapture = vi.fn()
    target.releasePointerCapture = vi.fn()
    target.hasPointerCapture = vi.fn(() => true)

    dragPaint.startDrag(createPointerEvent("pointerdown", target, 7, 5))
    dragPaint.moveDrag(createPointerEvent("pointermove", target, 7, 15))
    dragPaint.moveDrag(createPointerEvent("pointermove", target, 7, 25))
    dragPaint.endDrag(createPointerEvent("pointerup", target, 7, 25))

    expect(curScheduledEvent.value).toEqual({ col: 0, row: 0, numRows: 2 })
  })

  it("keeps the current timeslot aligned with the drag pointer while editing availability", () => {
    const { dragPaint, curTimeslot } = createAvailabilityDragPaint()
    const target = document.createElement("div")

    Object.defineProperty(target, "getBoundingClientRect", {
      configurable: true,
      value: () => ({
        left: 0,
        top: 0,
        width: 10,
        height: 10,
      }),
    })

    target.setPointerCapture = vi.fn()
    target.releasePointerCapture = vi.fn()
    target.hasPointerCapture = vi.fn(() => true)

    dragPaint.startDrag(createPointerEvent("pointerdown", target, 7, 5))
    expect(curTimeslot.value).toEqual({ row: 0, col: 0 })

    dragPaint.moveDrag(createPointerEvent("pointermove", target, 7, 35))
    expect(curTimeslot.value).toEqual({ row: 3, col: 0 })

    dragPaint.endDrag(createPointerEvent("pointerup", target, 7, 35))
    expect(curTimeslot.value).toEqual({ row: 3, col: 0 })
  })

  it("adds every exact UTC quarter-hour instant for a two-day specific-times selection", () => {
    const { dragPaint, dragStart, dragCur, tempTimes } =
      createSpecificTimesDragPaint()

    dragStart.value = { row: 0, col: 0 }
    dragCur.value = { row: 3, col: 1 }
    dragPaint.endDrag()

    expect([...tempTimes.value].map((time) => time.toString())).toEqual(
      Array.from({ length: 4 }, (_, row) =>
        ["2026-05-29", "2026-05-30"].map(
          (date) =>
            `${date}T00:${String(row * 15).padStart(2, "0")}:00+00:00[UTC]`,
        ),
      ).flat(),
    )
  })

  it("keeps the current timeslot aligned with the drag pointer while setting specific times", () => {
    const { dragPaint, curTimeslot } = createSpecificTimesDragPaint()
    const target = document.createElement("div")

    Object.defineProperty(target, "getBoundingClientRect", {
      configurable: true,
      value: () => ({
        left: 0,
        top: 0,
        width: 10,
        height: 10,
      }),
    })

    target.setPointerCapture = vi.fn()
    target.releasePointerCapture = vi.fn()
    target.hasPointerCapture = vi.fn(() => true)

    dragPaint.startDrag(createPointerEvent("pointerdown", target, 7, 5))
    expect(curTimeslot.value).toEqual({ row: 0, col: 0 })

    dragPaint.moveDrag(createPointerEvent("pointermove", target, 7, 35))
    expect(curTimeslot.value).toEqual({ row: 3, col: 0 })

    dragPaint.endDrag(createPointerEvent("pointerup", target, 7, 35))
    expect(curTimeslot.value).toEqual({ row: 3, col: 0 })
  })
})
