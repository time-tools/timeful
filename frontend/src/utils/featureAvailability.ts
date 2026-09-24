export interface SignInAvailabilityEnvironment {
  VITE_ENABLE_SIGN_IN?: string
}

export interface LandingAvailabilityEnvironment {
  VITE_ENABLE_RICH_LANDING?: string
}

export interface CalendarAutofillAvailabilityEnvironment {
  VITE_ENABLE_CALENDAR_AUTOFILL?: string
}

export interface CookieConsentAvailabilityEnvironment {
  VITE_ENABLE_COOKIE_CONSENT?: string
}

export interface LandingSignInAvailabilityEnvironment
  extends SignInAvailabilityEnvironment, LandingAvailabilityEnvironment {}

function isEnabled(rawValue: string | undefined): boolean {
  const value = rawValue?.trim().toLowerCase()

  if (!value) {
    return true
  }

  return value !== "false"
}

function isExplicitlyEnabled(rawValue: string | undefined): boolean {
  return rawValue?.trim().toLowerCase() === "true"
}

export function isSignInEnabled(
  env: SignInAvailabilityEnvironment = {},
): boolean {
  return isEnabled(env.VITE_ENABLE_SIGN_IN)
}

export function isRichLandingEnabled(
  env: LandingAvailabilityEnvironment = {},
): boolean {
  return isEnabled(env.VITE_ENABLE_RICH_LANDING)
}

export function isCalendarAutofillEnabled(
  env: CalendarAutofillAvailabilityEnvironment = {},
): boolean {
  return isEnabled(env.VITE_ENABLE_CALENDAR_AUTOFILL)
}

export function isCookieConsentEnabled(
  env: CookieConsentAvailabilityEnvironment = {},
): boolean {
  return isExplicitlyEnabled(env.VITE_ENABLE_COOKIE_CONSENT)
}

export function isLandingSignInEnabled(
  env: LandingSignInAvailabilityEnvironment = {},
): boolean {
  return isSignInEnabled(env) && isRichLandingEnabled(env)
}
