import { eventTypes, timeTypes, type TimeType } from "@/constants"
import type { Event } from "@/types"
import { Temporal } from "temporal-polyfill"
import type { Timezone } from "@/composables/schedule_overlap/types"
import type { PlainDate, ZonedDateTime } from "./temporalPrimitives"
import { getEventDateSeeds } from "./eventDateRules"
import { getSpecificTimesDayStarts } from "./scheduleDateRules"
import { getDateInTimezone, toZDT } from "./timezoneDateRules"

/** Returns a string representation of the given date, i.e. May 14th is "5/14". */
export const getDateString = (date: ZonedDateTime, utc = false): string => {
  const zdt = utc ? toZDT(date, "UTC") : toZDT(date)
  return `${String(zdt.month)}/${String(zdt.day)}`
}

/** Returns a string in the format "Mon, Sep 23, 10:00 AM - 12:00 PM PDT". */
export const getStartEndDateString = (
  startDate: ZonedDateTime,
  endDate: ZonedDateTime,
): string => {
  const start = toZDT(startDate)
  const end = toZDT(endDate)

  const startDay = start.toLocaleString("en-US", { weekday: "short" })
  const startMonth = start.toLocaleString("en-US", { month: "short" })
  const startDayOfMonth = start.toLocaleString("en-US", { day: "numeric" })
  const startTime = start.toLocaleString("en-US", {
    hour: "numeric",
    minute: "numeric",
  })
  const endTime = end.toLocaleString("en-US", {
    hour: "numeric",
    minute: "numeric",
    timeZoneName: "short",
  })

  return `${startDay}, ${startMonth} ${startDayOfMonth}, ${startTime} - ${endTime}`
}

/** Returns the shared clock format options for the given display time format. */
export const getTimeFormatOptions = (
  timeType: TimeType,
): Intl.DateTimeFormatOptions =>
  timeType === timeTypes.HOUR12
    ? { hour: "numeric", minute: "2-digit" }
    : { hour: "2-digit", minute: "2-digit", hour12: false }

/** Returns the shared full date options used for specific-date spans. */
export const specificDatesDateFormatOptions: Intl.DateTimeFormatOptions = {
  weekday: "short",
  month: "short",
  day: "numeric",
  year: "numeric",
}

/**
 * Returns the Event Occurrence Span text for the header, i.e.
 * "Tue, Sep 15, 2026 · 8:00 AM – 8:15 AM" or "Thu, May 28, 2026".
 */
export const getEventOccurrenceSpanString = ({
  startDate,
  endDate,
  timezone,
  timeType,
  daysOnly,
}: {
  startDate: ZonedDateTime
  endDate: ZonedDateTime
  timezone: Timezone
  timeType: TimeType
  daysOnly: boolean
}): string => {
  const start = getDateInTimezone(startDate, timezone)
  const end = getDateInTimezone(endDate, timezone)
  const startDateText = start.toLocaleString(
    "en-US",
    specificDatesDateFormatOptions,
  )

  if (daysOnly) {
    return startDateText
  }

  const timeFormat = getTimeFormatOptions(timeType)
  const startTimeText = start.toLocaleString("en-US", timeFormat)
  const endTimeText = end.toLocaleString("en-US", timeFormat)

  if (start.toPlainDate().equals(end.toPlainDate())) {
    return `${startDateText} \u00b7 ${startTimeText} \u2013 ${endTimeText}`
  }

  const endDateText = end.toLocaleString(
    "en-US",
    specificDatesDateFormatOptions,
  )

  return `${startDateText} \u00b7 ${startTimeText} \u2013 ${endDateText} \u00b7 ${endTimeText}`
}

/** Returns an ISO formatted date string. */
export const getISODateString = (date: ZonedDateTime, utc = false): string => {
  const zdt = utc ? toZDT(date, "UTC") : toZDT(date)
  return zdt.toPlainDate().toString()
}

const getPlainDateString = (date: PlainDate): string =>
  `${String(date.month)}/${String(date.day)}`

/** Returns a string representing date range from date1 to date2, i.e. "5/14 - 5/27". */
export const getDateRangeString = (
  date1: ZonedDateTime,
  date2: ZonedDateTime,
  utc = false,
): string => {
  const d1 = toZDT(date1, utc ? "UTC" : undefined)
  let d2 = toZDT(date2, utc ? "UTC" : undefined)

  if (d2.hour === 0 && d2.minute === 0 && d2.second === 0) {
    d2 = d2.subtract({ days: 1 })
  }

  return `${getDateString(d1, utc)} - ${getDateString(d2, utc)}`
}

/** Returns a string representing the date range for the provided event. */
export const getDateRangeStringForEvent = (
  event: Pick<
    Event,
    | "dates"
    | "daysOnly"
    | "timeSeed"
    | "type"
    | "activeSlots"
    | "eventTimezone"
    | "slotGeneration"
    | "timeIncrement"
    | "timedRecurrence"
  >,
  viewerTimezone?: Timezone,
): string => {
  if (!event.dates || event.dates.length === 0) return ""

  if (event.type === eventTypes.DOW || event.type === eventTypes.GROUP) {
    const dayAbbreviations = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
    const weeklyDays = event.timedRecurrence?.selectedDaysOfWeek?.length
      ? event.timedRecurrence.selectedDaysOfWeek
      : event.dates.map((date) => date.dayOfWeek)
    return weeklyDays
      .map((dayOfWeek) => dayAbbreviations[dayOfWeek % 7])
      .join(", ")
  }

  if (event.daysOnly) {
    return (
      `${getPlainDateString(event.dates[0])} - ` +
      getPlainDateString(event.dates[event.dates.length - 1])
    )
  }

  if (event.type === eventTypes.SPECIFIC_DATES) {
    if (viewerTimezone) {
      const viewerDays = getSpecificTimesDayStarts(
        getEventDateSeeds(event),
        viewerTimezone,
      )
      if (viewerDays.length > 0) {
        return (
          `${getPlainDateString(viewerDays[0].dateObject.toPlainDate())} - ` +
          getPlainDateString(
            viewerDays[viewerDays.length - 1].dateObject.toPlainDate(),
          )
        )
      }
    }

    const eventDateSeeds = getEventDateSeeds(event)
    if (eventDateSeeds.length === 0) {
      return ""
    }

    return (
      `${getPlainDateString(eventDateSeeds[0].toPlainDate())} - ` +
      getPlainDateString(
        eventDateSeeds[eventDateSeeds.length - 1].toPlainDate(),
      )
    )
  }

  return ""
}

/** Converts a timeNum (e.g. 13) to a timeText (e.g. "1 PM"). */
export const timeNumToTimeText = (timeNum: number, hour12 = true): string => {
  const hours = Math.floor(timeNum)
  const minutesDecimal = timeNum - hours
  const minutesString =
    minutesDecimal > 0
      ? `:${String(Math.floor(minutesDecimal * 60)).padStart(2, "0")}`
      : ""

  if (hour12) {
    if (timeNum >= 0 && timeNum < 1) return `12${minutesString} AM`
    if (timeNum < 12) return `${String(hours)}${minutesString} AM`
    if (timeNum >= 12 && timeNum < 13) return `12${minutesString} PM`
    return `${String(hours - 12)}${minutesString} PM`
  }

  return `${String(hours)}${minutesString.length > 0 ? minutesString : ":00"}`
}

/** Converts a timeNum (e.g. 9.5) to a timeString (e.g. 09:30:00). */
export const timeNumToTimeString = (timeNum: number): string => {
  const hours = Math.floor(timeNum)
  const minutesDecimal = timeNum - hours
  const paddedHours = String(hours).padStart(2, "0")
  const paddedMinutes = String(Math.floor(minutesDecimal * 60)).padStart(2, "0")

  return `${paddedHours}:${paddedMinutes}:00`
}

/** Returns the number of days in the given month. */
export const getDaysInMonth = (month: number, year: number): number => {
  return Temporal.PlainYearMonth.from({ year, month }).daysInMonth
}
