import { UAParser } from "ua-parser-js"

const browserAliases: Record<string, string> = {
  "Mobile Safari": "Safari",
  "Mobile Chrome": "Chrome",
  "Chrome WebView": "Chrome",
  "Microsoft Edge": "Edge",
}

const osAliases: Record<string, string> = {
  Ubuntu: "Linux",
  Debian: "Linux",
  "Chrome OS": "ChromeOS",
  "Mac OS": "macOS",
}

export function describeTargetBrowser(userAgent: string): string {
  if (!userAgent.trim()) return ""
  const { browser, os } = new UAParser(userAgent).getResult()
  const browserName = browser.name
    ? (browserAliases[browser.name] ?? browser.name)
    : ""
  const osName = os.name ? (osAliases[os.name] ?? os.name) : ""
  if (browserName && osName) return `${browserName} on ${osName}`
  return browserName || osName
}
