import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  integrations: [
    starlight({
      title: "Crona Docs",
      description: "Documentation for Crona.",
      social: [{ icon: "github", label: "GitHub", href: "https://github.com/webxsid/crona" }],
      sidebar: [{ label: "Documentation", items: [{ autogenerate: { directory: "." } }] }],
      customCss: ["./src/styles/custom.css"],
    }),
  ],
  site: "https://docs.crona.work",
});
