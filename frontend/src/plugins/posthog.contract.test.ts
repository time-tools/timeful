import { describe, expect, it } from "vitest"

describe("posthog-js dynamic import contract", () => {
  it("resolves the default client with the methods the plugin uses", async () => {
    const { default: client } = await import("posthog-js")

    expect(typeof client.init).toBe("function")
    expect(typeof client.capture).toBe("function")
    expect(typeof client.identify).toBe("function")
    expect(typeof client.reset).toBe("function")
    expect(typeof client.get_distinct_id).toBe("function")
  })
})
