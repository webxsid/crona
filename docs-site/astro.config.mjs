import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  integrations: [
    starlight({
      title: "Crona Docs",
      description: "Documentation for Crona.",
      social: [{ icon: "github", label: "GitHub", href: "https://github.com/webxsid/crona" }],
      sidebar: [
        {
          label: "Documentation",
          items: [
            {
              label: "Guides",
              items: [
                { label: "Getting Started", link: "/guides/getting-started/" },
                { label: "Concepts", link: "/guides/concepts/" },
                { label: "Features Overview", link: "/guides/features-overview/" },
                { label: "Screenshots and Walkthrough", link: "/guides/screenshots-and-walkthrough/" },
              ],
            },
            {
              label: "Workflows",
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
                { label: "Socket API", link: "/reference/socket-api/" },
                { label: "CLI and Local Engine", link: "/reference/cli-and-local-engine/" },
                { label: "TUI Keymap Reference", link: "/reference/tui-keymap-reference/" },
                { label: "Usage and Diagnostics", link: "/reference/usage-and-diagnostics/" },
              ],
            },
            {
              label: "Automation",
              items: [
                { label: "CLI Automation Patterns", link: "/automation/cli-automation-patterns/" },
                { label: "macOS Shortcuts Workflows", link: "/automation/macos-shortcuts-workflows/" },
              ],
            },
            {
              label: "Development",
              items: [
                { label: "Development", link: "/development/development/" },
                { label: "Contributing", link: "/development/contributing/" },
                { label: "Feature Design", link: "/development/feature-design/" },
              ],
            },
            {
              label: "Migration",
              items: [
                { label: "Migration", link: "/migration/" },
                { label: "Legacy to Homebrew", link: "/migration/legacy-to-brew/" },
                { label: "Legacy to Go", link: "/migration/legacy-to-go/" },
                { label: "Legacy to Scoop", link: "/migration/legacy-to-scoop/" },
              ],
            },
          ],
        },
      ],
      customCss: ["./src/styles/custom.css"],
    }),
  ],
  site: "https://docs.crona.work",
});
