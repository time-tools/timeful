import { randomUUID } from "node:crypto"
import { expect, type APIRequestContext } from "@playwright/test"
import { seedOtpChallenge } from "./database-inspect"

const otpCode = "123456"

// Seeds a valid OTP challenge so sign-in resolves the authoritative account
// contract the app uses.
export function seedOtpAccount(email: string): void {
  seedOtpChallenge(email, otpCode)
}

// Verifies the seeded OTP and returns the account _id, which is the external
// user identifier surfaced by AccountUser. The profile name is sent with the
// verification so sign-in journeys keep a deterministic display name.
export async function verifySignIn(
  request: APIRequestContext,
  email: string,
): Promise<string> {
  const response = await request.post("/api/auth/otp/verify", {
    data: {
      email,
      code: otpCode,
      timezoneOffset: 0,
      firstName: "E2E",
      lastName: "Deletion",
    },
  })
  expect(response.status()).toBe(200)
  const profile = (await response.json()) as { _id: string }
  return profile._id
}

export function newDeletableEmail(label: string): string {
  return `delete-${label}-${randomUUID()}@example.invalid`
}

export async function signInNewAccount(
  request: APIRequestContext,
  label: string,
): Promise<{ email: string; userId: string }> {
  const email = newDeletableEmail(label)
  seedOtpAccount(email)
  return { email, userId: await verifySignIn(request, email) }
}
