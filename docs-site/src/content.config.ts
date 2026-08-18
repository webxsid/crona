import { docsLoader } from "@astrojs/starlight/loaders";
import { docsSchema } from "@astrojs/starlight/schema";
import { defineCollection } from "astro:content";
import { z } from "astro/zod";
import { ExtendDocsSchema } from "lucode-starlight/schema";

const docs = defineCollection({
  loader: docsLoader(),
  schema: docsSchema({ extend: z.object({ hosted: z.literal(true) }).merge(ExtendDocsSchema) }),
});

export const collections = { docs };
