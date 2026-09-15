import { Temporal } from "temporal-polyfill"

import type { ZonedDateTime } from "./temporalPrimitives"

/** Converts a timestamp from a specified timezone to a ZonedDateTime on that timezone. */
export const convertToUTC = (
  dateTimeString: string,
  timezoneValue: string,
): Temporal.ZonedDateTime => {
  try {
    return Temporal.ZonedDateTime.from(`${dateTimeString}[${timezoneValue}]`, {
      overflow: "constrain",
    })
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    throw new Error(
      `Failed to convert timezone: ${message}. Timezone: ${timezoneValue}`,
      { cause: err },
    )
  }
}

/** Converts UTC slots into the selected display timezone. */
export const convertUTCSlotsToLocalISO = (
  slots: ZonedDateTime[] | undefined,
  timezoneValue?: string,
): Temporal.ZonedDateTime[] => {
  if (!slots) return []

  return slots.map((slot) => {
    if (!timezoneValue) {
      return slot
    }

    try {
      return slot.withTimeZone(timezoneValue)
    } catch {
      throw new Error(
        "Invalid temporal date provided to convertUTCSlotsToLocalISO",
      )
    }
  })
}
