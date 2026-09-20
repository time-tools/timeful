import { shallowMount } from "@vue/test-utils"
import { vi } from "vitest"
import { Temporal } from "temporal-polyfill"
import { eventTypes, timeTypes, UTC } from "@/constants"
import {
  states,
  type ScheduleOverlapEvent,
} from "@/composables/schedule_overlap/types"
import { createLocalStorageMock } from "@/test/localStorage"
import { ZdtMap } from "@/utils"
import ScheduleOverlap from "./ScheduleOverlap.vue"
import type {
  ScheduleOverlapEditingAvailabilityAsViewModel,
  ScheduleOverlapMobileOverlayViewModel,
  ScheduleOverlapRespondentsPanelViewModel,
  ScheduleOverlapSidebarViewModel,
  ScheduleOverlapToolRowViewModel,
} from "./scheduleOverlapViewModelContracts"

export type ScheduleOverlapWrapper = ReturnType<
  typeof shallowMount<typeof ScheduleOverlap>
>

export const zdt = (iso: string) =>
  Temporal.Instant.from(iso).toZonedDateTimeISO(UTC)

export const utcTimezone = {
  value: "UTC",
  offset: Temporal.Duration.from({ hours: 0 }),
  label: "UTC",
  gmtString: "GMT+0",
}

export const buildUtcSpecificTimes = (date: string, times: string[]) =>
  times.map((time) => zdt(`${date}T${time}Z`))

export const buildUtcQuarterHourSlots = (date: string) => {
  const midnight = Temporal.PlainDate.from(date).toZonedDateTime({
    timeZone: "UTC",
    plainTime: "00:00:00",
  })

  return Array.from({ length: 96 }, (_, index) =>
    midnight.add({ minutes: index * 15 }),
  )
}

export const buildCanonicalSpecificTimesEvent = ({
  name,
  dates,
}: {
  name: string
  dates: string[]
}): ScheduleOverlapEvent => ({
  ...buildScheduleOverlapProps().event,
  name,
  dates: dates.map((date) => Temporal.PlainDate.from(date)),
  timeSeed: zdt(`${dates[0]}T00:00:00Z`),
  startTime: Temporal.PlainTime.from("00:00"),
  duration: Temporal.Duration.from({ hours: 24 }),
  hasSpecificTimes: true,
  timeIncrement: Temporal.Duration.from({ minutes: 15 }),
  times: [],
  eventTimezone: "UTC",
  slotGeneration: {
    startTimeLocal: Temporal.PlainTime.from("00:00"),
    endTimeLocal: Temporal.PlainTime.from("00:00"),
    timeIncrement: Temporal.Duration.from({ minutes: 15 }),
  },
  timedRecurrence: {
    kind: "specific_dates",
    selectedDays: dates.map((date) => Temporal.PlainDate.from(date)),
    selectedDaysOfWeek: [],
    startOnMonday: false,
  },
})

export interface TimedGridPresentationForTest {
  days: { dateObject: Temporal.ZonedDateTime; isConsecutive?: boolean }[]
  renderedRows: {
    id: string
    kind: "timeslot" | "collapsed" | "filler" | "split-gap"
    startLabel?: string
    endLabel?: string
    timeText?: string
    height: number
  }[]
  splitTimes: {
    absoluteMinutes?: number
    displayedMinutes?: number
    text?: string
  }[][]
  timeAxisEndText?: string
  collapseDisabledTimes?: boolean
  actions?: { toggleCollapsedSpan?: (id: string) => void }
}

export const getTimedGridPresentation = (wrapper: ScheduleOverlapWrapper) =>
  wrapper
    .findComponent({ name: "ScheduleOverlapTimeGrid" })
    .props("timedGrid") as TimedGridPresentationForTest

export const stubResizeObserver = () => {
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {
        return undefined
      }

      disconnect() {
        return undefined
      }
    },
  )
}

export const stubScrollTo = () => {
  vi.stubGlobal("scrollTo", vi.fn())
}

export const scheduleOverlapGlobalStubs = {
  "v-btn": true,
  "v-card": true,
  "v-card-actions": true,
  "v-card-text": true,
  "v-card-title": true,
  "v-dialog": true,
  "v-expand-transition": true,
  "v-icon": true,
  "v-progress-circular": true,
  "v-spacer": true,
  "v-switch": true,
  "v-text-field": true,
  AlertText: true,
  AvailabilityTypeToggle: true,
  BufferTimeSwitch: true,
  CalendarAccounts: true,
  CalendarEventBlock: true,
  ColorLegend: true,
  ExpandableSection: true,
  GCalWeekSelector: true,
  RespondentsList: true,
  ScheduleOverlapMobileOverlay: true,
  SignUpBlocksList: true,
  SignUpCalendarBlock: true,
  SpecificTimesInstructions: true,
  ToolRow: true,
  Tooltip: {
    template: "<div><slot /></div>",
  },
  WorkingHoursToggle: true,
  ZigZag: true,
}

export const buildScheduleOverlapProps = () => ({
  event: {
    _id: "evt-1",
    name: "Overnight event",
    type: eventTypes.SPECIFIC_DATES,
    dates: [Temporal.PlainDate.from("2026-01-01")],
    timeSeed: zdt("2026-01-01T23:00:00Z"),
    startTime: Temporal.PlainTime.from("23:00"),
    duration: Temporal.Duration.from({ hours: 2 }),
    daysOnly: false,
  },
  alwaysShowCalendarEvents: true,
  sampleCalendarEventsByDay: [
    [
      {
        id: "cal-1",
        startDate: zdt("2026-01-02T00:00:00Z"),
        endDate: zdt("2026-01-02T00:30:00Z"),
        hoursOffset: Temporal.Duration.from({ hours: 1 }),
        hoursLength: Temporal.Duration.from({ minutes: 30 }),
        free: false,
        calendarId: "primary",
      },
    ],
  ],
})

export const buildRespondentsPanelViewModel =
  (): ScheduleOverlapRespondentsPanelViewModel => ({
    allAvailableNote: null,
    event: buildScheduleOverlapProps().event,
    eventId: "evt-1",
    curGuestId: "",
    ownedGuestResponseLookupKeys: [],
    guestResponseLookupKey: "",
    days: [zdt("2026-01-01T23:00:00Z")],
    times: [],
    curDate: zdt("2026-01-01T23:00:00Z"),
    curRespondent: "",
    curRespondents: [],
    curTimeslot: { dayIndex: 0, timeIndex: 0 },
    curTimeslotAvailability: {},
    curTimeslotInactive: false,
    curTimeslotCellState: null,
    curTimeslotCollapsed: false,
    respondents: [],
    parsedResponses: {},
    isOwner: false,
    isGroup: false,
    attendees: [],
    responsesFormatted: new ZdtMap<Set<string>>(),
    timezone: {
      value: UTC,
      offset: Temporal.Duration.from({ hours: 0 }),
      label: "UTC",
      gmtString: "GMT+0",
    },
    showCalendarEvents: true,
    showBestTimes: false,
    hideIfNeeded: false,
    collapseDisabledTimes: true,
    guestAddedAvailability: false,
    addingAvailabilityAsGuest: false,
  })

export const buildScheduleOverlapToolRowViewModel =
  (): ScheduleOverlapToolRowViewModel => ({
    event: buildScheduleOverlapProps().event,
    state: states.HEATMAP,
    states,
    actions: {
      updateCurTimezone: vi.fn(),
      resetCurTimezone: vi.fn(),
      updateTimeType: vi.fn(),
      updateMobileNumDays: vi.fn(),
      updateShowBestTimes: vi.fn(),
      updateHideIfNeeded: vi.fn(),
      updateCollapseDisabledTimes: vi.fn(),
      updateStartCalendarOnMonday: vi.fn(),
      updateWeekOffset: vi.fn(),
      scheduleEvent: vi.fn(),
      cancelScheduleEvent: vi.fn(),
      confirmScheduleEvent: vi.fn(),
    },
    curTimezone: buildRespondentsPanelViewModel().timezone,
    timezoneModified: false,
    startCalendarOnMonday: false,
    showBestTimes: false,
    hideIfNeeded: false,
    collapseDisabledTimes: true,
    isWeekly: false,
    calendarPermissionGranted: false,
    weekOffset: 0,
    timezoneReferenceDate: zdt("2026-01-01T12:00:00Z"),
    numResponses: 0,
    mobileNumDays: 3,
    showMobileNumDaysSwitch: true,
    allowScheduleEvent: true,
    timeType: timeTypes.HOUR12,
  })

export const buildScheduleOverlapSidebarViewModel =
  (): ScheduleOverlapSidebarViewModel => ({
    event: buildScheduleOverlapProps().event,
    state: states.EDIT_AVAILABILITY,
    isSignUp: false,
    isOwner: false,
    isGroup: false,
    isPhone: false,
    authUser: null,
    alreadyRespondedToSignUpForm: false,
    signUpBlocks: [],
    signUpBlocksToAdd: [],
    numTempTimes: 0,
    activeSlotsCount: 0,
    responseCount: 0,
    canCollapseHours: false,
    curGuestId: "",
    userHasResponded: false,
    addingAvailabilityAsGuest: false,
    canEditGuestName: false,
    newGuestName: "",
    editGuestNameDialog: false,
    editingAvailabilityAs: buildEditingAvailabilityAsViewModel(),
    availabilityType: "available",
    showOverlayAvailabilityToggle: false,
    overlayAvailability: false,
    calendarPermissionGranted: false,
    calendarEventsMap: {},
    sharedCalendarAccounts: {},
    showCalendarOptions: true,
    calendarOptionsDialog: false,
    bufferTime: { enabled: false, time: 0 },
    workingHours: { enabled: false, startTime: 9, endTime: 17 },
    curTimezone: buildRespondentsPanelViewModel().timezone,
    deleteAvailabilityDialog: false,
    rightSideWidth: "320px",
    hasNextPage: false,
    nextPage: () => {},
    toolRow: buildScheduleOverlapToolRowViewModel(),
    respondentsPanel: buildRespondentsPanelViewModel(),
  })

export const buildEditingAvailabilityAsViewModel =
  (): ScheduleOverlapEditingAvailabilityAsViewModel => ({
    visible: true,
    actionText: "Adding",
    actorName: "a guest",
    editableGuestName: null,
  })

export const buildScheduleOverlapMobileOverlayViewModel =
  (): ScheduleOverlapMobileOverlayViewModel => ({
    bottomOffset: "4rem",
    isGroup: false,
    editing: false,
    isSignUp: false,
    availabilityType: "available",
    isWeekly: false,
    calendarPermissionGranted: false,
    weekOffset: 0,
    event: buildRespondentsPanelViewModel().event,
    showCalendarOptions: true,
    showStickyRespondents: true,
    respondentsPanel: buildRespondentsPanelViewModel(),
    state: states.HEATMAP,
    numTempTimes: 0,
    editingAvailabilityAs: buildEditingAvailabilityAsViewModel(),
    newGuestName: "",
    editGuestNameDialog: false,
  })

export const installScheduleOverlapTestGlobals = () => {
  vi.stubGlobal(
    "localStorage",
    createLocalStorageMock({ timeType: timeTypes.HOUR12 }),
  )
  vi.stubGlobal(
    "fetch",
    vi.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        text: () => Promise.resolve("{}"),
      }),
    ),
  )
}

export const mountScheduleOverlap = (
  options: {
    props?: Record<string, unknown>
    global?: { stubs?: Record<string, unknown> }
  } = {},
): ScheduleOverlapWrapper => {
  const { props, global, ...mountOptions } = options

  return shallowMount<typeof ScheduleOverlap>(ScheduleOverlap, {
    ...mountOptions,
    props: {
      ...buildScheduleOverlapProps(),
      ...(props ?? {}),
    },
    global: {
      ...global,
      stubs: {
        ...scheduleOverlapGlobalStubs,
        ...(global?.stubs ?? {}),
      },
    },
  })
}
