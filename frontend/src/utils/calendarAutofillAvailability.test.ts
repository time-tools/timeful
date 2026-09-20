import { describe, expect, it } from "vitest"
import { isCalendarAutofillEnabled } from "./calendarAutofillAvailability"

describe("calendarAutofillAvailability", () => {
  it("enables calendar autofill by default", () => {
    expect(isCalendarAutofillEnabled()).toBe(true)
    expect(isCalendarAutofillEnabled({})).toBe(true)
    expect(
      isCalendarAutofillEnabled({ VITE_ENABLE_CALENDAR_AUTOFILL: "   " }),
    ).toBe(true)
  })

  it("disables calendar autofill when the env flag is false", () => {
    expect(
      isCalendarAutofillEnabled({ VITE_ENABLE_CALENDAR_AUTOFILL: "false" }),
    ).toBe(false)
    expect(
      isCalendarAutofillEnabled({ VITE_ENABLE_CALENDAR_AUTOFILL: " FALSE " }),
    ).toBe(false)
  })

  it("keeps calendar autofill enabled for other values", () => {
    expect(
      isCalendarAutofillEnabled({ VITE_ENABLE_CALENDAR_AUTOFILL: "true" }),
    ).toBe(true)
    expect(
      isCalendarAutofillEnabled({ VITE_ENABLE_CALENDAR_AUTOFILL: "0" }),
    ).toBe(true)
  })
})
