// @vitest-environment happy-dom

import { computed } from "vue"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { createLocalStorageMock } from "@/test/localStorage"
import { useScheduleOverlapPreferences } from "./useScheduleOverlapPreferences"

describe("useScheduleOverlapPreferences", () => {
  beforeEach(() => {
    vi.stubGlobal("localStorage", createLocalStorageMock())
  })

  it("clears the in-memory guest name when the trimmed input is blank", () => {
    const preferences = useScheduleOverlapPreferences({
      eventId: computed(() => "evt-1"),
    })

    preferences.setGuestName("   ")

    expect(preferences.guestName.value).toBeUndefined()
    expect(localStorage.getItem("evt-1.guestName")).toBeNull()
  })

  it("trims guest names before keeping them in local state", () => {
    const preferences = useScheduleOverlapPreferences({
      eventId: computed(() => "evt-1"),
    })

    preferences.setGuestName("  Ada  ")

    expect(preferences.guestName.value).toBe("Ada")
    expect(localStorage.getItem("evt-1.guestName")).toBe("Ada")
  })

  it("does not keep whitespace-only ownership names in local state", () => {
    const preferences = useScheduleOverlapPreferences({
      eventId: computed(() => "evt-1"),
    })

    preferences.setGuestName("Ada")
    preferences.setGuestOwnership({
      name: "   ",
      guestId: "guest_1",
      guestEditToken: "secret",
      guestEditPolicy: "protected",
      guestOwnershipMode: "token",
    })

    expect(preferences.guestName.value).toBe("Ada")
    expect(preferences.guestOwnership.value).toMatchObject({
      guestId: "guest_1",
      guestEditToken: "secret",
    })
  })

  it("trims ownership names before exposing them in local state", () => {
    const preferences = useScheduleOverlapPreferences({
      eventId: computed(() => "evt-1"),
    })

    preferences.setGuestOwnership({
      name: "  Ada  ",
      guestOwnershipMode: "legacy",
    })

    expect(preferences.guestName.value).toBe("Ada")
    expect(preferences.guestOwnership.value?.name).toBe("Ada")
  })

  it("exposes only responses the visitor can edit", () => {
    const event = computed(() => ({
      eventVisitorId: "vp_visitor",
      responses: {
        rp_ada: { name: "Ada", publicId: "rp_ada", canEdit: true },
        rp_other: { name: "Grace", publicId: "rp_other", canEdit: false },
      },
    }))
    const preferences = useScheduleOverlapPreferences({
      eventId: computed(() => "evt-1"),
      event,
    })

    expect(preferences.ownedGuestResponses.value).toEqual([
      { lookupKey: "rp_ada", name: "Ada", lastUsedAt: 0 },
    ])
    expect(preferences.getOwnedGuestOwnership("rp_other")).toBeUndefined()

    preferences.selectGuestOwnership("rp_ada")

    expect(preferences.guestOwnership.value?.lookupKey).toBe("rp_ada")
  })
})
