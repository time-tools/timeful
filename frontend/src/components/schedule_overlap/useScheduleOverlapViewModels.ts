import { computed, type ComputedRef, type Ref } from "vue"
import { Temporal } from "temporal-polyfill"
import type { AvailabilityType } from "@/constants"
import type {
  CalendarEventsByDay,
  CalendarEventsMap,
  DayItem,
  FetchedResponse,
  MonthDayItem,
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
import type {
  ScheduleOverlapDaysOnlyGridActions,
  ScheduleOverlapDaysOnlyGridViewModel,
  ScheduleOverlapEditingAvailabilityAsViewModel,
  ScheduleOverlapMobileOverlayViewModel,
  ScheduleOverlapRespondentsPanelViewModel,
  ScheduleOverlapSidebarViewModel,
  ScheduleOverlapTimeGridActions,
  ScheduleOverlapTimeGridViewModel,
  ScheduleOverlapToolRowActions,
  ScheduleOverlapToolRowViewModel,
} from "./scheduleOverlapViewModelContracts"
import type {
  ClassStyle,
  RenderedOverlayAvailabilityFragment,
} from "./scheduleOverlapRendering"
import type { User } from "@/types"
import type { ZdtMap } from "@/utils"
import type { useAvailabilityData } from "@/composables/schedule_overlap/useAvailabilityData"
import type { useCalendarEvents } from "@/composables/schedule_overlap/useCalendarEvents"
import type { useCalendarGrid } from "@/composables/schedule_overlap/useCalendarGrid"
import type { useEventScheduling } from "@/composables/schedule_overlap/useEventScheduling"
import type { useSignUpForm } from "@/composables/schedule_overlap/useSignUpForm"
import type { useScheduleOverlapUI } from "@/composables/schedule_overlap/useScheduleOverlapUI"
import type { useOwnedTimezone } from "@/composables/timezone/useOwnedTimezone"
import type { useTimedGridPresentation } from "./useTimedGridPresentation"

interface UseScheduleOverlapViewModelsFlatOptions {
  event: ComputedRef<ScheduleOverlapEvent>
  state: Ref<ScheduleOverlapState>
  states: Record<string, ScheduleOverlapState>
  isSignUp: ComputedRef<boolean>
  isOwner: ComputedRef<boolean>
  isGroup: ComputedRef<boolean>
  isPhone: ComputedRef<boolean>
  authUser: ComputedRef<{
    firstName?: string
    lastName?: string
    calendarAccounts?: Record<string, CalendarAccountEntry>
  } | null>
  alreadyRespondedToSignUpForm: Ref<boolean>
  signUpBlocksByDay: Ref<SignUpBlockLite[][]>
  signUpBlocksToAddByDay: Ref<SignUpBlockLite[][]>
  tempTimes: Ref<{ size: number }>
  curGuestId: ComputedRef<string>
  ownedGuestResponseLookupKeys: ComputedRef<string[]>
  guestResponseLookupKey: ComputedRef<string>
  userHasResponded: Ref<boolean>
  addingAvailabilityAsGuest: ComputedRef<boolean>
  canEditGuestName: ComputedRef<boolean>
  newGuestName: Ref<string>
  editGuestNameDialog: Ref<boolean>
  availabilityType: Ref<AvailabilityType>
  showOverlayAvailabilityToggle: ComputedRef<boolean>
  overlayAvailability: Ref<boolean>
  calendarPermissionGranted: ComputedRef<boolean>
  calendarEventsMap: ComputedRef<CalendarEventsMap>
  sharedCalendarAccounts: Ref<Record<string, CalendarAccountEntry>>
  showCalendarOptions: ComputedRef<boolean>
  calendarOptionsDialog: Ref<boolean>
  bufferTime: Ref<{ enabled: boolean; time: number }>
  workingHours: Ref<{ enabled: boolean; startTime: number; endTime: number }>
  curTimezone: Ref<Timezone>
  deleteAvailabilityDialog: Ref<boolean>
  rightSideWidth: ComputedRef<string>
  allDays: Ref<unknown[]>
  times: Ref<TimeItem[]>
  getDateFromRowCol: (
    row: number,
    col: number,
  ) => Temporal.ZonedDateTime | Temporal.PlainDate | null
  curTimeslot: Ref<RowCol>
  curRespondent: Ref<string>
  curRespondents: Ref<string[]>
  curTimeslotAvailability: Ref<Record<string, boolean>>
  curTimeslotInactive: Ref<boolean>
  curTimeslotCellState: Ref<TimedCellState | null>
  curTimeslotCollapsed: Ref<boolean>
  respondents: ComputedRef<User[]>
  parsedResponses: ComputedRef<ParsedResponses>
  attendees: ComputedRef<{ email: string; declined?: boolean }[] | undefined>
  responsesFormatted: Ref<ZdtMap<Set<string>>>
  showCalendarEvents: Ref<boolean>
  showBestTimes: Ref<boolean>
  hideIfNeeded: Ref<boolean>
  collapseDisabledTimesPreference: Ref<boolean>
  collapseDisabledTimes: ComputedRef<boolean>
  guestAddedAvailability: ComputedRef<boolean>
  editing: ComputedRef<boolean>
  isWeekly: Ref<boolean>
  weekOffset: ComputedRef<number>
  delayedShowStickyRespondents: Ref<boolean>
  toolRowActions: ComputedRef<ScheduleOverlapToolRowActions>
  timezoneModified: Ref<boolean>
  startCalendarOnMonday: Ref<boolean>
  allowDrag: ComputedRef<boolean>
  timezoneReferenceDate: Ref<Temporal.ZonedDateTime>
  mobileNumDays: Ref<number>
  allowScheduleEvent: Ref<boolean>
  timeType: Ref<string>
  daysOnlyGridActions: ComputedRef<ScheduleOverlapDaysOnlyGridActions>
  curMonthText: Ref<string>
  hasPrevPage: Ref<boolean>
  hasNextPage: Ref<boolean>
  daysOfWeek: Ref<string[]>
  monthDays: Ref<MonthDayItem[]>
  dayTimeslotClassStyle: ComputedRef<ClassStyle[]>
  dayTimeslotVon: ComputedRef<Record<string, () => void>[]>
  calendarOnly: ComputedRef<boolean>
  timedGridActions: ComputedRef<ScheduleOverlapTimeGridActions>
  splitTimes: Ref<TimeItem[][]>
  timeslotHeight: Ref<number>
  renderedRows: ComputedRef<RenderedTimeGridRow[]>
  timeAxisEndText: ComputedRef<string | undefined>
  days: Ref<DayItem[]>
  isSpecificDates: Ref<boolean>
  sampleCalendarEventsByDay: ComputedRef<CalendarEventsByDay>
  showLoader: ComputedRef<boolean>
  loadingCalendarEvents: ComputedRef<boolean>
  alwaysShowCalendarEvents: ComputedRef<boolean>
  calendarEventsByDay: Ref<CalendarEventsByDay>
  page: Ref<number>
  maxDaysPerPage: Ref<number>
  dragStart: Ref<RowCol | null>
  curScheduledEvent: Ref<ScheduledEvent | null>
  savedScheduledEvent: ComputedRef<ScheduledEvent | null>
  scheduledEventStyle: Ref<Record<string, string>>
  scheduledEventStyles: ComputedRef<Record<string, string>[]>
  signUpBlockBeingDraggedStyle: Ref<Record<string, string>>
  newSignUpBlockName: Ref<string>
  overlaidAvailability: ComputedRef<RenderedOverlayAvailabilityFragment[][]>
  timeslotClassStyle: ComputedRef<ClassStyle[]>
  timeslotVon: ComputedRef<Record<string, () => void>[]>
  noEventNames: ComputedRef<boolean>
  max: Ref<number>
  fetchedResponses: Ref<Record<string, FetchedResponse | undefined>>
  loadingResponsesLoading: ComputedRef<boolean>
  getRenderedTimeBlockStyles: (block: {
    hoursOffset?: Temporal.Duration
    hoursLength?: Temporal.Duration
  }) => Record<string, string>[]
  getRenderedTimeBlockStyle: (block: {
    hoursOffset?: Temporal.Duration
    hoursLength?: Temporal.Duration
  }) => Record<string, string>
  getSignUpBlockStyle: (block: SignUpBlockLite) => Record<string, string>
}

interface UseScheduleOverlapViewModelsOptions {
  event: ComputedRef<ScheduleOverlapEvent>
  state: Ref<ScheduleOverlapState>
  states: Record<string, ScheduleOverlapState>
  isSignUp: ComputedRef<boolean>
  isOwner: ComputedRef<boolean>
  isGroup: ComputedRef<boolean>
  isPhone: ComputedRef<boolean>
  authUser: UseScheduleOverlapViewModelsFlatOptions["authUser"]
  props: Pick<
    UseScheduleOverlapViewModelsFlatOptions,
    | "curGuestId"
    | "addingAvailabilityAsGuest"
    | "calendarPermissionGranted"
    | "calendarEventsMap"
    | "calendarOnly"
    | "loadingCalendarEvents"
    | "sampleCalendarEventsByDay"
    | "alwaysShowCalendarEvents"
    | "noEventNames"
    | "weekOffset"
  >
  derived: Pick<
    UseScheduleOverlapViewModelsFlatOptions,
    "showCalendarOptions" | "showLoader" | "attendees" | "collapseDisabledTimes"
  >
  rendering: Pick<
    UseScheduleOverlapViewModelsFlatOptions,
    "loadingResponsesLoading" | "getSignUpBlockStyle"
  >
  preferences: Pick<
    UseScheduleOverlapViewModelsFlatOptions,
    "showBestTimes" | "collapseDisabledTimesPreference"
  >
  guest: Pick<
    UseScheduleOverlapViewModelsFlatOptions,
    | "ownedGuestResponseLookupKeys"
    | "guestResponseLookupKey"
    | "guestAddedAvailability"
  >
  timezone: ReturnType<typeof useOwnedTimezone>
  grid: ReturnType<typeof useCalendarGrid>
  calendarEvents: ReturnType<typeof useCalendarEvents>
  availability: ReturnType<typeof useAvailabilityData>
  signUpForm: ReturnType<typeof useSignUpForm>
  ui: ReturnType<typeof useScheduleOverlapUI>
  scheduling: ReturnType<typeof useEventScheduling>
  presentation: ReturnType<typeof useTimedGridPresentation>
  dragStart: Ref<RowCol | null>
  actions: Pick<
    UseScheduleOverlapViewModelsFlatOptions,
    "toolRowActions" | "daysOnlyGridActions" | "timedGridActions"
  >
}

interface AllAvailableNoteInput {
  editingAvailability: boolean
  loadingResponses: boolean
  fetchedResponsesCount: number
  max: number
  respondentsLength: number
  daysOnly: boolean
  isGroup: boolean
}

export const buildAllAvailableNote = (
  input: AllAvailableNoteInput,
): string | null => {
  if (input.editingAvailability) return null
  if (input.loadingResponses) return null
  if (input.fetchedResponsesCount === 0) return null
  if (input.max >= input.respondentsLength) return null
  const unit = input.daysOnly ? "day" : "time"
  const noun = input.isGroup ? "members" : "respondents"
  return `Note: There's no ${unit} when all ${input.respondentsLength} ${noun} are available.`
}

export function useScheduleOverlapViewModels(
  input: UseScheduleOverlapViewModelsOptions,
) {
  const opts: UseScheduleOverlapViewModelsFlatOptions = {
    ...input.props,
    ...input.derived,
    ...input.rendering,
    ...input.preferences,
    ...input.guest,
    ...input.grid,
    ...input.calendarEvents,
    ...input.availability,
    ...input.signUpForm,
    ...input.ui,
    ...input.scheduling,
    ...input.presentation,
    ...input.actions,
    event: input.event,
    state: input.state,
    states: input.states,
    isSignUp: input.isSignUp,
    isOwner: input.isOwner,
    isGroup: input.isGroup,
    isPhone: input.isPhone,
    authUser: input.authUser,
    dragStart: input.dragStart,
    curTimezone: input.timezone.timezone,
    timezoneModified: input.timezone.modified,
  }
  const respondentsPanel = computed<ScheduleOverlapRespondentsPanelViewModel>(
    () => ({
      allAvailableNote: buildAllAvailableNote({
        editingAvailability: opts.state.value === opts.states.EDIT_AVAILABILITY,
        loadingResponses: opts.loadingResponsesLoading.value,
        fetchedResponsesCount: Object.keys(opts.fetchedResponses.value).length,
        max: opts.max.value,
        respondentsLength: opts.respondents.value.length,
        daysOnly: Boolean(opts.event.value.daysOnly),
        isGroup: opts.isGroup.value,
      }),
      event: opts.event.value,
      eventId: opts.event.value._id ?? "",
      curGuestId: opts.curGuestId.value,
      ownedGuestResponseLookupKeys: opts.ownedGuestResponseLookupKeys.value,
      guestResponseLookupKey: opts.guestResponseLookupKey.value,
      days: opts.allDays.value,
      times: opts.times.value,
      curDate: (() => {
        const curDate = opts.getDateFromRowCol(
          opts.curTimeslot.value.row,
          opts.curTimeslot.value.col,
        )
        return curDate instanceof Temporal.ZonedDateTime ? curDate : undefined
      })(),
      curRespondent: opts.curRespondent.value,
      curRespondents: opts.curRespondents.value,
      curTimeslot: {
        dayIndex: opts.curTimeslot.value.col,
        timeIndex: opts.curTimeslot.value.row,
      },
      curTimeslotAvailability: opts.curTimeslotAvailability.value,
      curTimeslotInactive: opts.curTimeslotInactive.value,
      curTimeslotCellState: opts.curTimeslotCellState.value,
      curTimeslotCollapsed: opts.curTimeslotCollapsed.value,
      respondents: opts.respondents.value,
      parsedResponses: opts.parsedResponses.value,
      isOwner: opts.isOwner.value,
      isGroup: opts.isGroup.value,
      attendees: opts.attendees.value,
      responsesFormatted: opts.responsesFormatted.value,
      timezone: opts.curTimezone.value,
      showCalendarEvents: opts.showCalendarEvents.value,
      showBestTimes: opts.showBestTimes.value,
      hideIfNeeded: opts.hideIfNeeded.value,
      collapseDisabledTimes: opts.collapseDisabledTimes.value,
      guestAddedAvailability: opts.guestAddedAvailability.value,
      addingAvailabilityAsGuest: opts.addingAvailabilityAsGuest.value,
    }),
  )

  const editingAvailabilityAs =
    computed<ScheduleOverlapEditingAvailabilityAsViewModel>(() => {
      const curGuestId = opts.curGuestId.value
      const guestName = curGuestId
        ? (opts.event.value.responses?.[curGuestId]?.name ?? curGuestId)
        : ""
      const authUser = opts.authUser.value
      return {
        visible: !(
          opts.calendarPermissionGranted.value &&
          !opts.event.value.daysOnly &&
          !opts.addingAvailabilityAsGuest.value
        ),
        actionText:
          (opts.userHasResponded.value &&
            !opts.addingAvailabilityAsGuest.value) ||
          curGuestId
            ? "Editing"
            : "Adding",
        actorName: (() => {
          if (authUser && !opts.addingAvailabilityAsGuest.value) {
            return `${authUser.firstName ?? ""} ${authUser.lastName ?? ""}`.trim()
          }
          if (curGuestId.length > 0) {
            return guestName
          }
          return "a guest"
        })(),
        editableGuestName:
          curGuestId && opts.canEditGuestName.value ? guestName : null,
      }
    })

  const sidebarViewModel = computed<ScheduleOverlapSidebarViewModel>(() => ({
    event: opts.event.value,
    state: opts.state.value,
    isSignUp: opts.isSignUp.value,
    isOwner: opts.isOwner.value,
    isGroup: opts.isGroup.value,
    isPhone: opts.isPhone.value,
    authUser: opts.authUser.value,
    alreadyRespondedToSignUpForm: opts.alreadyRespondedToSignUpForm.value,
    signUpBlocks: opts.signUpBlocksByDay.value.flat(),
    signUpBlocksToAdd: opts.signUpBlocksToAddByDay.value.flat(),
    numTempTimes: opts.tempTimes.value.size,
    activeSlotsCount: opts.event.value.daysOnly
      ? (opts.event.value.dates ?? []).length
      : (opts.event.value.activeSlots ?? opts.event.value.times ?? []).length,
    responseCount: opts.respondents.value.length,
    canCollapseHours: (() => {
      const state = opts.state.value
      return (
        !opts.event.value.daysOnly &&
        state !== opts.states.EDIT_SIGN_UP_BLOCKS &&
        state !== opts.states.SET_SPECIFIC_TIMES &&
        opts.collapseDisabledTimesPreference.value
      )
    })(),
    curGuestId: opts.curGuestId.value,
    userHasResponded: opts.userHasResponded.value,
    addingAvailabilityAsGuest: opts.addingAvailabilityAsGuest.value,
    canEditGuestName: opts.canEditGuestName.value,
    newGuestName: opts.newGuestName.value,
    editGuestNameDialog: opts.editGuestNameDialog.value,
    editingAvailabilityAs: editingAvailabilityAs.value,
    availabilityType: opts.availabilityType.value,
    showOverlayAvailabilityToggle: opts.showOverlayAvailabilityToggle.value,
    overlayAvailability: opts.overlayAvailability.value,
    calendarPermissionGranted: opts.calendarPermissionGranted.value,
    calendarEventsMap: opts.calendarEventsMap.value,
    sharedCalendarAccounts: opts.sharedCalendarAccounts.value,
    showCalendarOptions: opts.showCalendarOptions.value,
    calendarOptionsDialog: opts.calendarOptionsDialog.value,
    bufferTime: opts.bufferTime.value,
    workingHours: opts.workingHours.value,
    curTimezone: opts.curTimezone.value,
    deleteAvailabilityDialog: opts.deleteAvailabilityDialog.value,
    rightSideWidth: opts.rightSideWidth.value,
    hasNextPage: opts.hasNextPage.value,
    nextPage: opts.timedGridActions.value.nextPage,
    toolRow: toolRowViewModel.value,
    respondentsPanel: respondentsPanel.value,
  }))

  const mobileOverlayViewModel =
    computed<ScheduleOverlapMobileOverlayViewModel>(() => ({
      bottomOffset: "4rem",
      isGroup: opts.isGroup.value,
      editing: opts.editing.value,
      isSignUp: opts.isSignUp.value,
      availabilityType: opts.availabilityType.value,
      isWeekly: opts.isWeekly.value,
      calendarPermissionGranted: opts.calendarPermissionGranted.value,
      weekOffset: opts.weekOffset.value,
      event: opts.event.value,
      showCalendarOptions: opts.showCalendarOptions.value,
      showStickyRespondents: opts.delayedShowStickyRespondents.value,
      respondentsPanel: respondentsPanel.value,
      state: opts.state.value,
      numTempTimes: opts.tempTimes.value.size,
      editingAvailabilityAs: editingAvailabilityAs.value,
      newGuestName: opts.newGuestName.value,
      editGuestNameDialog: opts.editGuestNameDialog.value,
    }))

  const toolRowViewModel = computed<ScheduleOverlapToolRowViewModel>(() => ({
    event: opts.event.value,
    state: opts.state.value,
    states: opts.states,
    actions: opts.toolRowActions.value,
    curTimezone: opts.curTimezone.value,
    timezoneModified: opts.timezoneModified.value,
    startCalendarOnMonday: opts.startCalendarOnMonday.value,
    showBestTimes: opts.showBestTimes.value,
    hideIfNeeded: opts.hideIfNeeded.value,
    collapseDisabledTimes: opts.collapseDisabledTimes.value,
    isWeekly: opts.isWeekly.value,
    calendarPermissionGranted: opts.calendarPermissionGranted.value,
    weekOffset: opts.weekOffset.value,
    timezoneReferenceDate: opts.timezoneReferenceDate.value,
    numResponses: opts.respondents.value.length,
    mobileNumDays: opts.mobileNumDays.value,
    showMobileNumDaysSwitch: opts.allDays.value.length > 3,
    allowScheduleEvent: opts.allowScheduleEvent.value,
    timeType: opts.timeType.value,
  }))

  const daysOnlyGridViewModel = computed<ScheduleOverlapDaysOnlyGridViewModel>(
    () => ({
      event: opts.event.value,
      actions: opts.daysOnlyGridActions.value,
      curMonthText: opts.curMonthText.value,
      hasPrevPage: opts.hasPrevPage.value,
      hasNextPage: opts.hasNextPage.value,
      daysOfWeek: opts.daysOfWeek.value,
      monthDays: opts.monthDays.value,
      dayTimeslotClassStyle: opts.dayTimeslotClassStyle.value,
      dayTimeslotVon: opts.dayTimeslotVon.value,
      allowDrag: opts.allowDrag.value,
      isPhone: opts.isPhone.value,
      calendarOnly: opts.calendarOnly.value,
      toolRow: toolRowViewModel.value,
    }),
  )

  const timedGridViewModel = computed<ScheduleOverlapTimeGridViewModel>(() => ({
    event: opts.event.value,
    actions: opts.timedGridActions.value,
    calendarOnly: opts.calendarOnly.value,
    hasPrevPage: opts.hasPrevPage.value,
    hasNextPage: opts.hasNextPage.value,
    splitTimes: opts.splitTimes.value,
    times: opts.times.value,
    renderedRows: opts.renderedRows.value,
    timeAxisEndText: opts.timeAxisEndText.value,
    timeslotHeight: opts.timeslotHeight.value,
    days: opts.days.value,
    isSpecificDates: opts.isSpecificDates.value,
    isGroup: opts.isGroup.value,
    sampleCalendarEventsByDay: opts.sampleCalendarEventsByDay.value,
    showLoader: opts.showLoader.value,
    loadingCalendarEvents: opts.loadingCalendarEvents.value,
    editing: opts.editing.value,
    alwaysShowCalendarEvents: opts.alwaysShowCalendarEvents.value,
    showCalendarEvents: opts.showCalendarEvents.value,
    calendarEventsByDay: opts.calendarEventsByDay.value,
    state: opts.state.value,
    states: opts.states,
    page: opts.page.value,
    maxDaysPerPage: opts.maxDaysPerPage.value,
    dragStart: opts.dragStart.value,
    curScheduledEvent: opts.curScheduledEvent.value,
    savedScheduledEvent: opts.savedScheduledEvent.value,
    scheduledEventStyle: opts.scheduledEventStyle.value,
    scheduledEventStyles: opts.scheduledEventStyles.value,
    signUpBlockBeingDraggedStyle: opts.signUpBlockBeingDraggedStyle.value,
    newSignUpBlockName: opts.newSignUpBlockName.value,
    isSignUp: opts.isSignUp.value,
    signUpBlocksByDay: opts.signUpBlocksByDay.value,
    signUpBlocksToAddByDay: opts.signUpBlocksToAddByDay.value,
    overlayAvailability: opts.overlayAvailability.value,
    overlaidAvailability: opts.overlaidAvailability.value,
    timeslotClassStyle: opts.timeslotClassStyle.value,
    timeslotVon: opts.timeslotVon.value,
    noEventNames: opts.noEventNames.value,
    isPhone: opts.isPhone.value,
    loadingResponsesLoading: opts.loadingResponsesLoading.value,
    allowDrag: opts.allowDrag.value,
    toolRow: toolRowViewModel.value,
    getRenderedTimeBlockStyles: opts.getRenderedTimeBlockStyles,
    getRenderedTimeBlockStyle: opts.getRenderedTimeBlockStyle,
    getSignUpBlockStyle: opts.getSignUpBlockStyle,
  }))

  return {
    respondentsPanel,
    sidebarViewModel,
    mobileOverlayViewModel,
    toolRowViewModel,
    daysOnlyGridViewModel,
    timedGridViewModel,
  }
}
