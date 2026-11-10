import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 120_000,
  expect: {
    timeout: 15_000,
  },
  fullyParallel: false,
  workers: 1,
  reporter: [["list"]],
  outputDir: "../../artifacts/playwright-results",
  use: {
    ...devices["Desktop Chrome"],
    browserName: "chromium",
    channel: "chrome",
    headless: true,
    ignoreHTTPSErrors: true,
    launchOptions: {
      args: ["--ignore-certificate-errors"],
    },
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    trace: "off",
  },
});
