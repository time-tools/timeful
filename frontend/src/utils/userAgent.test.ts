import { describe, expect, it } from "vitest"
import { describeTargetBrowser } from "./userAgent"

const firefoxLinux =
  "Mozilla/5.0 (X11; Linux x86_64; rv:141.0) Gecko/20100101 Firefox/141.0"
const chromeWindows =
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
const edgeWindows =
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0"
const safariMac =
  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15"
const chromeAndroid =
  "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Mobile Safari/537.36"
const safariIPhone =
  "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1"

describe("describeTargetBrowser", () => {
  it("names the browser and operating system", () => {
    expect(describeTargetBrowser(firefoxLinux)).toBe("Firefox on Linux")
    expect(describeTargetBrowser(chromeWindows)).toBe("Chrome on Windows")
    expect(describeTargetBrowser(edgeWindows)).toBe("Edge on Windows")
    expect(describeTargetBrowser(safariMac)).toBe("Safari on macOS")
  })

  it("uses shared browser names for mobile clients", () => {
    expect(describeTargetBrowser(chromeAndroid)).toBe("Chrome on Android")
    expect(describeTargetBrowser(safariIPhone)).toBe("Safari on iOS")
  })

  it("returns the identified half when only one side is known", () => {
    expect(describeTargetBrowser("Mozilla/5.0 Firefox/141.0")).toBe("Firefox")
    expect(describeTargetBrowser("Mozilla/5.0 (X11; Linux x86_64)")).toBe(
      "Linux",
    )
  })

  it("returns an empty description for empty or unrecognized values", () => {
    expect(describeTargetBrowser("")).toBe("")
    expect(describeTargetBrowser("   ")).toBe("")
    expect(describeTargetBrowser("not a browser")).toBe("")
  })
})
