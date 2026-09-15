// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { createLocalStorageMock } from "@/test/localStorage"

const POSTHOG_DISTINCT_ID_KEY = "timeful.posthogDistinctId"
const API_KEY = "test-posthog-key"

const { mockClient } = vi.hoisted(() => ({
  mockClient: {
    init: vi.fn(),
    capture: vi.fn(),
    identify: vi.fn(),
    reset: vi.fn(),
    get_distinct_id: vi.fn(() => "client-distinct-id"),
  },
}))

vi.mock("posthog-js", () => ({ default: mockClient }))

function importPlugin() {
  return import("./posthog")
}

describe("posthog plugin", () => {
  beforeEach(() => {
    vi.resetModules()
    vi.clearAllMocks()
    vi.stubGlobal("localStorage", createLocalStorageMock())
    vi.stubEnv("VITE_POSTHOG_API_KEY", API_KEY)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.unstubAllEnvs()
  })

  it("lazily initializes with the configured key and privacy-preserving options", async () => {
    const { posthog } = await importPlugin()

    posthog.capture("test_event")

    await vi.waitFor(() => {
      expect(mockClient.init).toHaveBeenCalledWith(API_KEY, {
        capture_pageview: false,
        autocapture: false,
      })
    })
  })

  it("forwards queued capture calls to the loaded client", async () => {
    const { posthog } = await importPlugin()

    posthog.capture("test_event", { answer: 42 })

    await vi.waitFor(() => {
      expect(mockClient.capture).toHaveBeenCalledWith("test_event", {
        answer: 42,
      })
    })
  })

  it("forwards identify calls with a distinct id", async () => {
    const { posthog } = await importPlugin()

    posthog.identify("user-1", { email: "user@example.com" })

    await vi.waitFor(() => {
      expect(mockClient.identify).toHaveBeenCalledWith(
        "user-1",
        {
          email: "user@example.com",
        },
        undefined,
      )
    })
  })

  it("skips identify and client loading when the distinct id is missing", async () => {
    const { posthog } = await importPlugin()

    posthog.identify(undefined, { email: "user@example.com" })

    await Promise.resolve()

    expect(mockClient.identify).not.toHaveBeenCalled()
    expect(mockClient.init).not.toHaveBeenCalled()
  })

  it("clears the stored fallback id and resets the client", async () => {
    localStorage.setItem(POSTHOG_DISTINCT_ID_KEY, "fallback-id")
    const { posthog } = await importPlugin()

    posthog.reset()

    await vi.waitFor(() => {
      expect(mockClient.reset).toHaveBeenCalled()
    })
    expect(localStorage.getItem(POSTHOG_DISTINCT_ID_KEY)).toBeNull()
  })

  it("returns a stable generated fallback distinct id before the client loads", async () => {
    vi.stubEnv("VITE_POSTHOG_API_KEY", "")
    const { posthog } = await importPlugin()

    const first = posthog.get_distinct_id()
    const second = posthog.get_distinct_id()

    expect(first).toBeTruthy()
    expect(second).toBe(first)
    expect(mockClient.init).not.toHaveBeenCalled()
  })

  it("returns the client distinct id once the client has loaded", async () => {
    const { posthog } = await importPlugin()

    await vi.waitFor(() => {
      expect(posthog.get_distinct_id()).toBe("client-distinct-id")
    })
  })
})
