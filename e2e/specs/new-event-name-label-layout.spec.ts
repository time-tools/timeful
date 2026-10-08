import { expect, test } from "@playwright/test"

test("keeps the floating Event name label inside the New event form scroll area", async ({
  page,
}) => {
  await page.goto("/")
  await page.getByRole("button", { name: "Create event", exact: true }).click()

  const nameInput = page
    .getByRole("dialog")
    .getByRole("textbox", { name: "Event name (required)" })
  await expect(nameInput).toBeFocused()

  // The outlined field draws its floating label across the top border, above
  // the field box, so the form's overflow container must leave room for it.
  await expect
    .poll(() =>
      nameInput.evaluate((input) => {
        const label = input
          .closest(".v-input")
          ?.querySelector(".v-field-label--floating")
        const scrollArea = input.closest(".v-card-text")
        if (!label || !scrollArea) {
          return null
        }

        return Math.round(
          label.getBoundingClientRect().top -
            scrollArea.getBoundingClientRect().top,
        )
      }),
    )
    .toBeGreaterThanOrEqual(0)
})
