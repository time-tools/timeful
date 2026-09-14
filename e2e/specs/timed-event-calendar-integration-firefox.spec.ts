import { expect, test, type Page } from "@playwright/test"
import { Temporal } from "temporal-polyfill"
import { signInNewAccount } from "../helpers/account-auth"
import { databaseExec, databaseScalar } from "../helpers/database-inspect"

const mockCalendarEmail = "calendar-mock@example.invalid"
const mockCalendarType = "google"
const mockSubCalendarId = "work"

// The mock provider serves one fixed account email, so every test in this file
// connects the same calendar_key to a different account. Each assertion joins
// through platform_identities.id, which is the account identifier the sign-in
// session carries, so parallel tests cannot read or mutate another account's
// rows.
function countCalendarAccounts(userId: string): string {
  return databaseScalar(
    `SELECT count(*) FROM calendar_accounts a
     JOIN platform_identities p ON p.id = a.platform_identity_id
     WHERE p.id = '${userId}'`,
  )
}

function countCalendarCredentials(userId: string): string {
  return databaseScalar(
    `SELECT count(*) FROM calendar_account_credentials c
     JOIN calendar_accounts a ON a.id = c.calendar_account_id
     JOIN platform_identities p ON p.id = a.platform_identity_id
     WHERE p.id = '${userId}'`,
  )
}

function countSubCalendars(userId: string): string {
  return databaseScalar(
    `SELECT count(*) FROM calendar_sub_calendars s
     JOIN calendar_accounts a ON a.id = s.calendar_account_id
     JOIN platform_identities p ON p.id = a.platform_identity_id
     WHERE p.id = '${userId}'`,
  )
}

function encryptedAccessToken(userId: string): string {
  return databaseScalar(
    `SELECT c.oauth_access_token_ciphertext FROM calendar_account_credentials c
     JOIN calendar_accounts a ON a.id = c.calendar_account_id
     JOIN platform_identities p ON p.id = a.platform_identity_id
     WHERE p.id = '${userId}'`,
  )
}

function accessTokenExpiryIsPast(userId: string): string {
  return databaseScalar(
    `SELECT (c.oauth_access_token_expires_at < clock_timestamp())::text FROM calendar_account_credentials c
     JOIN calendar_accounts a ON a.id = c.calendar_account_id
     JOIN platform_identities p ON p.id = a.platform_identity_id
     WHERE p.id = '${userId}'`,
  )
}

function subCalendarEnabled(userId: string, subCalendarId: string): string {
  return databaseScalar(
    `SELECT s.enabled::text FROM calendar_sub_calendars s
     JOIN calendar_accounts a ON a.id = s.calendar_account_id
     JOIN platform_identities p ON p.id = a.platform_identity_id
     WHERE p.id = '${userId}' AND s.sub_calendar_id = '${subCalendarId}'`,
  )
}

// Adds a Google connection through the running server. server-test routes the
// OAuth exchange and calendar-list calls to the test-only calendar-mock provider
// through the TEST_-prefixed endpoint overrides, so no live provider is reached.
async function connectMockGoogleCalendar(page: Page): Promise<void> {
  const response = await page.request.post(
    "/api/user/add-google-calendar-account",
    { data: { code: "mock-authorization-code", scope: "openid" } },
  )
  expect(response.status()).toBe(200)
}

function calendarEventsURL(): string {
  const now = Temporal.Now.instant()
  return `/api/user/calendars?timeMin=${encodeURIComponent(now.toString())}&timeMax=${encodeURIComponent(now.add({ hours: 1 }).toString())}`
}

test("connecting a Google calendar persists encrypted credentials", async ({
  page,
}) => {
  const { userId } = await signInNewAccount(page.request, "calendar-connect")

  await test.step("the connection is added through the running server", async () => {
    await connectMockGoogleCalendar(page)
  })

  await test.step("the database stores the account and encrypted credentials", () => {
    expect(countCalendarAccounts(userId)).toBe("1")
    expect(countCalendarCredentials(userId)).toBe("1")
    expect(countSubCalendars(userId)).toBe("2")

    const ciphertext = encryptedAccessToken(userId)
    expect(ciphertext).not.toHaveLength(0)
    expect(ciphertext).not.toContain("mock-access-token")
  })

  await test.step("the connection renders in the Firefox settings page", async () => {
    await page.goto("/settings", { waitUntil: "domcontentloaded" })
    await expect(
      page.getByText(mockCalendarEmail, { exact: true }),
    ).toBeVisible()
  })
})

test("an expired OAuth2 connection refreshes its encrypted access token", async ({
  page,
}) => {
  const { userId } = await signInNewAccount(page.request, "calendar-refresh")
  await connectMockGoogleCalendar(page)

  const before = encryptedAccessToken(userId)
  databaseExec(
    `UPDATE calendar_account_credentials c
     SET oauth_access_token_expires_at = clock_timestamp() - interval '1 hour'
     FROM calendar_accounts a
     JOIN platform_identities p ON p.id = a.platform_identity_id
     WHERE c.calendar_account_id = a.id AND p.id = '${userId}'`,
  )
  expect(accessTokenExpiryIsPast(userId)).toBe("true")

  await test.step("loading calendars refreshes the expired token", async () => {
    const response = await page.request.get(calendarEventsURL())
    expect(response.status()).toBe(200)
  })

  await test.step("the database stores a new encrypted token and future expiry", () => {
    const after = encryptedAccessToken(userId)
    expect(after).not.toHaveLength(0)
    expect(after).not.toBe(before)
    expect(accessTokenExpiryIsPast(userId)).toBe("false")
  })
})

test("toggling a sub-calendar persists its enabled state", async ({ page }) => {
  const { userId } = await signInNewAccount(page.request, "calendar-toggle")
  await connectMockGoogleCalendar(page)

  expect(subCalendarEnabled(userId, mockSubCalendarId)).toBe("false")

  await test.step("enabling the sub-calendar persists true", async () => {
    const response = await page.request.post("/api/user/toggle-sub-calendar", {
      data: {
        email: mockCalendarEmail,
        calendarType: mockCalendarType,
        subCalendarId: mockSubCalendarId,
        enabled: true,
      },
    })
    expect(response.status()).toBe(200)
    expect(subCalendarEnabled(userId, mockSubCalendarId)).toBe("true")
  })

  await test.step("disabling the sub-calendar persists false", async () => {
    const response = await page.request.post("/api/user/toggle-sub-calendar", {
      data: {
        email: mockCalendarEmail,
        calendarType: mockCalendarType,
        subCalendarId: mockSubCalendarId,
        enabled: false,
      },
    })
    expect(response.status()).toBe(200)
    expect(subCalendarEnabled(userId, mockSubCalendarId)).toBe("false")
  })
})

test("removing a calendar connection cascades its credentials and sub-calendars", async ({
  page,
}) => {
  const { userId } = await signInNewAccount(page.request, "calendar-remove")
  await connectMockGoogleCalendar(page)
  expect(countCalendarAccounts(userId)).toBe("1")
  expect(countCalendarCredentials(userId)).toBe("1")
  expect(countSubCalendars(userId)).toBe("2")

  await test.step("removing the connection deletes the account", async () => {
    const response = await page.request.delete(
      "/api/user/remove-calendar-account",
      {
        data: {
          email: mockCalendarEmail,
          calendarType: mockCalendarType,
        },
      },
    )
    expect(response.status()).toBe(200)
  })

  await test.step("the account and its cascaded rows are gone", () => {
    expect(countCalendarAccounts(userId)).toBe("0")
    expect(countCalendarCredentials(userId)).toBe("0")
    expect(countSubCalendars(userId)).toBe("0")
  })
})
