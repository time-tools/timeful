import { describe, expect, it } from "vitest"
import { isCookieConsentEnabled } from "./cookieConsentAvailability"

describe("cookieConsentAvailability", () => {
  it("is disabled by default", () => {
    expect(isCookieConsentEnabled()).toBe(false)
    expect(isCookieConsentEnabled({})).toBe(false)
    expect(isCookieConsentEnabled({ VITE_ENABLE_COOKIE_CONSENT: "   " })).toBe(
      false,
    )
  })

  it("enables only an explicit true value", () => {
    expect(isCookieConsentEnabled({ VITE_ENABLE_COOKIE_CONSENT: "true" })).toBe(
      true,
    )
    expect(
      isCookieConsentEnabled({ VITE_ENABLE_COOKIE_CONSENT: " TRUE " }),
    ).toBe(true)
    expect(
      isCookieConsentEnabled({ VITE_ENABLE_COOKIE_CONSENT: "false" }),
    ).toBe(false)
    expect(isCookieConsentEnabled({ VITE_ENABLE_COOKIE_CONSENT: "0" })).toBe(
      false,
    )
  })
})
