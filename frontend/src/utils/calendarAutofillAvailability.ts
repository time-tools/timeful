import { isCalendarAutofillEnabled } from "./featureAvailability"

export type { CalendarAutofillAvailabilityEnvironment } from "./featureAvailability"
export { isCalendarAutofillEnabled } from "./featureAvailability"

export const calendarAutofillEnabled = isCalendarAutofillEnabled(
  import.meta.env,
)
