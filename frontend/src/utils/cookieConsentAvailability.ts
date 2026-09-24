import { isCookieConsentEnabled } from "./featureAvailability"

export type { CookieConsentAvailabilityEnvironment } from "./featureAvailability"
export { isCookieConsentEnabled } from "./featureAvailability"

export const cookieConsentEnabled = isCookieConsentEnabled(import.meta.env)
