import type { EventDraft } from "@/composables/event/types"
import type { Temporal } from "temporal-polyfill"
import type { components } from "./api"
import type {
  RawCalendarAccount,
  RawCalendarEvent,
  RawCalendarOptions,
  RawEvent,
  RawResponse,
  RawSignUpBlock,
  RawSignUpResponse,
  RawUser,
} from "./transport"

type Schemas = components["schemas"]

export type User = Omit<RawUser, "calendarAccounts" | "calendarOptions"> & {
  calendarAccounts?: Record<string, CalendarAccount>
  calendarOptions?: CalendarOptions
}

export type Event = Omit<
  RawEvent,
  | "dates"
  | "times"
  | "duration"
  | "timeIncrement"
  | "activeSlots"
  | "slotGeneration"
  | "timedRecurrence"
  | "scheduledEvent"
  | "responses"
  | "signUpBlocks"
  | "signUpResponses"
> & {
  dates?: Temporal.PlainDate[]
  timeSeed?: Temporal.ZonedDateTime
  times?: Temporal.ZonedDateTime[]
  startTime?: Temporal.PlainTime
  endTime?: Temporal.PlainTime
  duration?: Temporal.Duration
  timeIncrement?: Temporal.Duration
  activeSlots?: Temporal.ZonedDateTime[]
  eventTimezone?: string
  slotGeneration?: {
    startTimeLocal?: Temporal.PlainTime
    endTimeLocal?: Temporal.PlainTime
    timeIncrement?: Temporal.Duration
  }
  timedRecurrence?: {
    kind?: "specific_dates" | "weekly"
    selectedDays?: Temporal.PlainDate[]
    selectedDaysOfWeek?: number[]
    startOnMonday?: boolean
  }
  scheduledEvent?: CalendarEvent
  responses?: Record<string, Response>
  signUpBlocks?: SignUpBlock[]
  signUpResponses?: Record<string, SignUpResponse>
}

export interface Folder {
  _id?: Schemas["routes.FolderResponse"]["_id"]
  color?: Schemas["routes.FolderResponse"]["color"]
  eventIds?: Schemas["routes.FolderResponse"]["eventIds"]
  isDeleted?: Schemas["routes.FolderResponse"]["isDeleted"]
  name?: Schemas["routes.FolderResponse"]["name"]
  userId?: Schemas["routes.FolderResponse"]["userId"]
}

export type Response = Omit<
  RawResponse,
  | "availability"
  | "ifNeeded"
  | "manualAvailability"
  | "calendarOptions"
  | "user"
> & {
  availability?: Temporal.ZonedDateTime[]
  ifNeeded?: Temporal.ZonedDateTime[]
  manualAvailability?: Record<string, Temporal.ZonedDateTime[]>
  calendarOptions?: CalendarOptions
  user?: User
}

export type SignUpBlock = Omit<RawSignUpBlock, "startDate" | "endDate"> & {
  startDate?: Temporal.ZonedDateTime
  endDate?: Temporal.ZonedDateTime
}

export type SignUpResponse = Omit<RawSignUpResponse, "user"> & {
  user?: User
}

export type SignUpBlockWithResponses = SignUpBlock & {
  responses?: SignUpResponse[]
}

export type CalendarAccount = Omit<RawCalendarAccount, "subCalendars"> & {
  subCalendars?: Record<string, SubCalendar>
}

export type CalendarEvent = Omit<RawCalendarEvent, "startDate" | "endDate"> & {
  startDate?: Temporal.ZonedDateTime
  endDate?: Temporal.ZonedDateTime
}

export type CalendarOptions = Omit<
  RawCalendarOptions,
  "bufferTime" | "workingHours"
> & {
  bufferTime?: BufferTimeOptions
  workingHours?: WorkingHoursOptions
}

export interface SubCalendar {
  enabled?: Schemas["models.SubCalendar"]["enabled"]
  name?: Schemas["models.SubCalendar"]["name"]
}

export interface Location {
  city?: Schemas["models.Location"]["city"]
  country_code?: Schemas["models.Location"]["country_code"]
  country_name?: Schemas["models.Location"]["country_name"]
  latitude?: Schemas["models.Location"]["latitude"]
  longitude?: Schemas["models.Location"]["longitude"]
  postal?: Schemas["models.Location"]["postal"]
  state?: Schemas["models.Location"]["state"]
}

export interface Remindee {
  email?: Schemas["models.Remindee"]["email"]
  responded?: Schemas["models.Remindee"]["responded"]
}

// Group event reads inject the attendee list alongside the event payload.
export interface Attendee {
  _id?: string
  declined?: boolean
  email?: string
  eventId?: string
}

export interface BufferTimeOptions {
  enabled?: Schemas["models.BufferTimeOptions"]["enabled"]
  time?: Schemas["models.BufferTimeOptions"]["time"]
}

export interface WorkingHoursOptions {
  enabled?: Schemas["models.WorkingHoursOptions"]["enabled"]
  endTime?: Schemas["models.WorkingHoursOptions"]["endTime"]
  startTime?: Schemas["models.WorkingHoursOptions"]["startTime"]
}

export interface NewDialogOptions {
  show: boolean
  contactsPayload: EventDraft
  openNewGroup: boolean
  eventOnly: boolean
  folderId: string | null
}
