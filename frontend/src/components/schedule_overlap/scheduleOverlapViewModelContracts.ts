import type { AvailabilityType } from "@/constants"
import type {
  CalendarEventsMap,
  CalendarEventsByDay,
  DayItem,
  MonthDayItem,
  NormalizedCalendarEvent,
  ParsedResponses,
  RenderedTimeGridRow,
  RowCol,
  ScheduleOverlapEvent,
  ScheduleOverlapState,
  ScheduledEvent,
  SignUpBlockLite,
  TimeItem,
  TimedCellState,
  Timezone,
} from "@/composables/schedule_overlap/types"
import type { CalendarAccountEntry } from "@/components/settings/CalendarAccounts.vue"
import type { Temporal } from "temporal-polyfill"
import type { User } from "@/types"
import type { ZdtMap } from "@/utils"
import type {
  ClassStyle,
  OverlaidAvailabilityBlock,
  RenderedOverlayAvailabilityFragment,
} from "./scheduleOverlapRendering"

export interface ScheduleOverlapToolRowActions {
  updateCurTimezone: (value: Timezone) => void
  resetCurTimezone: () => void
  updateTimeType: (value: string | number) => void
  updateMobileNumDays: (value: number) => void
  updateShowBestTimes: (value: boolean) => void
  updateHideIfNeeded: (value: boolean) => void
  updateCollapseDisabledTimes: (value: boolean) => void
  updateStartCalendarOnMonday: (value: boolean) => void
  updateWeekOffset: (value: number) => void
  scheduleEvent: (e?: MouseEvent) => void
  cancelScheduleEvent: (e?: MouseEvent) => void
  confirmScheduleEvent: (
    destination?: "timeful" | "google" | "outlook" | boolean,
  ) => void
}

export interface ScheduleOverlapDaysOnlyGridActions {
  prevPage: (e?: Event) => void
  nextPage: (e?: Event) => void
  startDrag: (e: PointerEvent | MouseEvent) => void
  moveDrag: (e: PointerEvent | MouseEvent) => void
  endDrag: (e?: PointerEvent | MouseEvent) => void
  resetCurTimeslot: () => void
}

export interface ScheduleOverlapTimeGridActions {
  prevPage: (e?: Event) => void
  nextPage: (e?: Event) => void
  calendarScroll: (event: Event) => void
  startDrag: (e: PointerEvent | MouseEvent) => void
  moveDrag: (e: PointerEvent | MouseEvent) => void
  endDrag: (e?: PointerEvent | MouseEvent) => void
  resetCurTimeslot: () => void
  signUpForBlock: (block: SignUpBlockLite) => void
  toggleCollapsedSpan: (id: string) => void
  markCollapsedRowInactive: () => void
  markSplitGapOutside: () => void
  clickSplitGapOutside: () => void
}

export interface ScheduleOverlapRespondentsPanelViewModel {
  allAvailableNote: string | null
  event: ScheduleOverlapEvent
  eventId: string
  curGuestId: string
  ownedGuestResponseLookupKeys: string[]
  guestResponseLookupKey: string
  days: unknown[]
  times: unknown[]
  curDate?: Temporal.ZonedDateTime
  curRespondent: string
  curRespondents: string[]
  curTimeslot: { dayIndex: number; timeIndex: number }
  curTimeslotAvailability: Record<string, boolean>
  curTimeslotInactive: boolean
  curTimeslotCellState: TimedCellState | null
  curTimeslotCollapsed: boolean
  respondents: User[]
  parsedResponses: ParsedResponses
  isOwner: boolean
  isGroup: boolean
  attendees?: { email: string; declined?: boolean }[]
  responsesFormatted: ZdtMap<Set<string>>
  timezone: Timezone
  showCalendarEvents: boolean
  showBestTimes: boolean
  hideIfNeeded: boolean
  collapseDisabledTimes: boolean
  guestAddedAvailability: boolean
  addingAvailabilityAsGuest: boolean
}

export interface ScheduleOverlapEditingAvailabilityAsViewModel {
  visible: boolean
  actionText: "Editing" | "Adding"
  actorName: string
  editableGuestName: string | null
}

export interface ScheduleOverlapSidebarViewModel {
  event: ScheduleOverlapEvent
  state: ScheduleOverlapState
  isSignUp: boolean
  isOwner: boolean
  isGroup: boolean
  isPhone: boolean
  authUser: {
    firstName?: string
    lastName?: string
    calendarAccounts?: Record<string, CalendarAccountEntry>
  } | null
  alreadyRespondedToSignUpForm: boolean
  signUpBlocks: SignUpBlockLite[]
  signUpBlocksToAdd: SignUpBlockLite[]
  numTempTimes: number
  activeSlotsCount: number
  responseCount: number
  canCollapseHours: boolean
  curGuestId?: string
  userHasResponded: boolean
  addingAvailabilityAsGuest: boolean
  canEditGuestName: boolean
  newGuestName: string
  editGuestNameDialog: boolean
  editingAvailabilityAs: ScheduleOverlapEditingAvailabilityAsViewModel
  availabilityType: AvailabilityType
  showOverlayAvailabilityToggle: boolean
  overlayAvailability: boolean
  calendarPermissionGranted: boolean
  calendarEventsMap: CalendarEventsMap
  sharedCalendarAccounts: Record<string, CalendarAccountEntry>
  showCalendarOptions: boolean
  calendarOptionsDialog: boolean
  bufferTime: { enabled: boolean; time: number }
  workingHours: { enabled: boolean; startTime: number; endTime: number }
  curTimezone: Timezone
  deleteAvailabilityDialog: boolean
  rightSideWidth: string
  hasNextPage: boolean
  nextPage: () => void
  toolRow: ScheduleOverlapToolRowViewModel
  respondentsPanel: ScheduleOverlapRespondentsPanelViewModel
}

export interface ScheduleOverlapMobileOverlayViewModel {
  bottomOffset: string
  isGroup: boolean
  editing: boolean
  isSignUp: boolean
  availabilityType: AvailabilityType
  isWeekly: boolean
  calendarPermissionGranted: boolean
  weekOffset: number
  event: ScheduleOverlapEvent
  showCalendarOptions: boolean
  showStickyRespondents: boolean
  respondentsPanel: ScheduleOverlapRespondentsPanelViewModel
  state: ScheduleOverlapState
  numTempTimes: number
  editingAvailabilityAs: ScheduleOverlapEditingAvailabilityAsViewModel
  newGuestName: string
  editGuestNameDialog: boolean
}

export interface ScheduleOverlapToolRowViewModel {
  event: ScheduleOverlapEvent
  state: ScheduleOverlapState
  states: Record<string, ScheduleOverlapState>
  actions: ScheduleOverlapToolRowActions
  curTimezone: Timezone
  timezoneModified: boolean
  startCalendarOnMonday: boolean
  showBestTimes: boolean
  hideIfNeeded: boolean
  collapseDisabledTimes: boolean
  isWeekly: boolean
  calendarPermissionGranted: boolean
  weekOffset: number
  timezoneReferenceDate: Temporal.ZonedDateTime
  numResponses: number
  mobileNumDays: number
  showMobileNumDaysSwitch: boolean
  allowScheduleEvent: boolean
  timeType: string
}

export interface ScheduleOverlapDaysOnlyGridViewModel {
  event: ScheduleOverlapEvent
  actions: ScheduleOverlapDaysOnlyGridActions
  curMonthText: string
  hasPrevPage: boolean
  hasNextPage: boolean
  daysOfWeek: string[]
  monthDays: MonthDayItem[]
  dayTimeslotClassStyle: ClassStyle[]
  dayTimeslotVon: Record<string, () => void>[]
  allowDrag: boolean
  isPhone: boolean
  calendarOnly: boolean
  toolRow: ScheduleOverlapToolRowViewModel
}

export interface ScheduleOverlapTimeGridViewModel {
  event: ScheduleOverlapEvent
  actions: ScheduleOverlapTimeGridActions
  calendarOnly: boolean
  hasPrevPage: boolean
  hasNextPage: boolean
  splitTimes: TimeItem[][]
  times: TimeItem[]
  renderedRows: RenderedTimeGridRow[]
  timeAxisEndText?: string
  timeslotHeight: number
  days: DayItem[]
  isSpecificDates: boolean
  isGroup: boolean
  sampleCalendarEventsByDay: unknown
  showLoader: boolean
  loadingCalendarEvents: boolean
  editing: boolean
  alwaysShowCalendarEvents: boolean
  showCalendarEvents: boolean
  calendarEventsByDay: CalendarEventsByDay
  state: ScheduleOverlapState
  states: Record<string, ScheduleOverlapState>
  page: number
  maxDaysPerPage: number
  dragStart: RowCol | null
  curScheduledEvent: ScheduledEvent | null
  savedScheduledEvent?: ScheduledEvent | null
  scheduledEventStyle: Record<string, string>
  scheduledEventStyles: Record<string, string>[]
  signUpBlockBeingDraggedStyle: Record<string, string>
  newSignUpBlockName: string
  isSignUp: boolean
  signUpBlocksByDay: SignUpBlockLite[][]
  signUpBlocksToAddByDay: SignUpBlockLite[][]
  overlayAvailability: boolean
  overlaidAvailability: RenderedOverlayAvailabilityFragment[][]
  timeslotClassStyle: ClassStyle[]
  timeslotVon: Record<string, () => void>[]
  noEventNames: boolean
  isPhone: boolean
  loadingResponsesLoading: boolean
  allowDrag: boolean
  toolRow: ScheduleOverlapToolRowViewModel
  getRenderedTimeBlockStyles: (
    block: NormalizedCalendarEvent | OverlaidAvailabilityBlock,
  ) => Record<string, string>[]
  getRenderedTimeBlockStyle: (
    block: NormalizedCalendarEvent | OverlaidAvailabilityBlock,
  ) => Record<string, string>
  getSignUpBlockStyle: (block: SignUpBlockLite) => Record<string, string>
}
