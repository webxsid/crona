import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import { ion } from "starlight-ion-theme";

export default defineConfig({
  integrations: [
    starlight({
      title: "Crona Docs",
      description: "Documentation for Crona.",
      social: [
        { icon: "github", label: "GitHub", href: "https://github.com/webxsid/crona" },
        { icon: "threads", label: "Threads", href: "https://www.threads.com/@crona.work" },
        { icon: "blueSky", label: "Bluesky", href: "https://bsky.app/profile/crona.work" },
      ],
      plugins: [ion()],
      sidebar: [
        {
          label: "Documentation",
          items: [
            {
              label: "Start here",
              items: [
                { label: "Install", link: "/install/" },
                { label: "Getting Started", link: "/guides/getting-started/" },
                { label: "Beta Channel", link: "/guides/beta-channel/", badge: { text: "Beta" } },
                { label: "Concepts", link: "/guides/concepts/" },
                { label: "Features Overview", link: "/guides/features-overview/" },
                { label: "Screenshots and Walkthrough", link: "/guides/screenshots-and-walkthrough/" },
              ],
            },
            {
              label: "Use Crona",
              items: [
                { label: "Issues and Planning", link: "/workflows/issues-and-planning/" },
                { label: "Habits", link: "/workflows/habits/" },
                { label: "Focus Sessions", link: "/workflows/focus-sessions/" },
                { label: "Check-Ins and Wellbeing", link: "/workflows/check-ins-and-wellbeing/" },
                { label: "Alerts and Reminders", link: "/workflows/alerts-and-reminders/" },
                { label: "Exports and Reports", link: "/workflows/exports-and-reports/" },
                { label: "Calendar and File Automation", link: "/workflows/calendar-and-file-automation/" },
                { label: "macOS Companion", link: "/workflows/macos-companion/" },
              ],
            },
            {
              label: "Reference",
              items: [
                { label: "CLI and Local Engine", link: "/reference/cli-and-local-engine/" },
                { label: "TUI Keymap Reference", link: "/reference/tui-keymap-reference/", badge: { text: "Beta" } },
                { label: "Usage and Diagnostics", link: "/reference/usage-and-diagnostics/" },
              ],
            },
            {
              label: "API Reference",
              items: [
                { label: "Overview", link: "/api/", badge: { text: "Beta" } },
                { label: "Transport and Envelopes", link: "/api/transport/", badge: { text: "Beta" } },
                { label: "Runtime and Operations", link: "/api/runtime/" },
                { label: "Work Management", link: "/api/work-management/" },
                { label: "Focus and Wellbeing", link: "/api/focus-and-wellbeing/", badge: { text: "Beta" } },
                { label: "Exports and Settings", link: "/api/exports-and-settings/" },
                { label: "Events", link: "/api/events/", badge: { text: "Beta" } },
              ],
            },
            {
              label: "Automate",
              items: [
                { label: "CLI Automation Patterns", link: "/automation/cli-automation-patterns/" },
                { label: "macOS Shortcuts Workflows", link: "/automation/macos-shortcuts-workflows/" },
              ],
            },
          ],
        },
      ],
    }),
  ],
  site: "https://docs.crona.work",
});
