import { describe, expect, it } from "vitest"
import { eventPublicId } from "./eventIdentity"

describe("eventPublicId", () => {
  it("returns the canonical public identifier", () => {
    expect(eventPublicId({ _id: "7Q2M4XKP", shortId: "7Q2M4XKP" })).toBe(
      "7Q2M4XKP",
    )
  })

  it("falls back to the short identifier when _id is absent", () => {
    expect(eventPublicId({ shortId: "7Q2M4XKP" })).toBe("7Q2M4XKP")
  })

  it("returns an empty string when no identifier exists", () => {
    expect(eventPublicId({})).toBe("")
  })
})
