import { initSchema } from "./schema";

export async function initDb() {
  try {
    await initSchema();
    console.log("Database schema initialized successfully.");
  } catch (error) {
    console.error("Error initializing database schema:", error);
    throw error;
  }
}
