import { describe, expect, it } from "vitest"
import {
  getGuestNameValidationMessage,
  getResponseDisplayName,
  hasGuestName,
  normalizeGuestName,
  validateGuestName,
} from "./guestName"

describe("guestName boundary", () => {
  it("trims valid guest names and strips formatting characters", () => {
    expect(normalizeGuestName("  A\u200bda\u0301  ")).toBe("Adá")
  })

  it("rejects blank or formatting-only guest names", () => {
    expect(validateGuestName("   ").code).toBe("required")
    expect(validateGuestName("\u200b\u200b").code).toBe("invalidFormatting")
    expect(hasGuestName("   ")).toBe(false)
  })

  it("rejects account-id-like and overlength guest names", () => {
    expect(validateGuestName("0197c9a2-6c3f-7b8e-9f01-2f3a4b5c6d7e").code).toBe(
      "accountIdLike",
    )
    expect(validateGuestName("507f1f77bcf86cd799439011").normalizedName).toBe(
      "507f1f77bcf86cd799439011",
    )
    expect(validateGuestName("a".repeat(101)).code).toBe("tooLong")
    expect(validateGuestName("e\u0301".repeat(101)).code).toBe("tooLong")
  })

  it("returns specific guest-name validation messages", () => {
    expect(getGuestNameValidationMessage("required")).toBe(
      "Name must be non-empty",
    )
    expect(getGuestNameValidationMessage("invalidFormatting")).toContain(
      "formatting",
    )
    expect(getGuestNameValidationMessage("accountIdLike")).toContain(
      "account ID",
    )
    expect(getGuestNameValidationMessage("tooLong")).toContain("100")
  })

  it("formats response display names without undefined fallbacks", () => {
    expect(
      getResponseDisplayName({
        user: {
          firstName: "Ada",
        },
      }),
    ).toBe("Ada")
    expect(
      getResponseDisplayName({
        name: "Guest Ada",
      }),
    ).toBe("Guest Ada")
  })
})
