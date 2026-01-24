import type { FastifyInstance } from "fastify";

export function registerAuth(
  app: FastifyInstance,
  token: string
) {
  app.addHook("preHandler", async (req) => {
    // Health is public (used for discovery)
    if (req.originalUrl === "/health") return;

    // if (req.originalUrl === "/events") {
    // for the events url the token will be passed as a query parameter 'auth'
    // const authToken = (req.query as { auth: string | undefined }).auth;
    // if (!authToken || authToken !== token) {
    //   throw new Error("Unauthorized");
    // }
    //   return;
    // }

    const header = req.headers.authorization;
    if (!header) {
      throw new Error("Unauthorized");
    }

    const value = header.replace("Bearer ", "");
    if (value !== token) {
      throw new Error("Unauthorized");
    }
  });
}
