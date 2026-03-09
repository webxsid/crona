import { bootstrapKernel } from "./bootstrap";
import { logError, logInfo } from "./logger";

process.on("uncaughtException", (err) => {
  console.error("Uncaught exception:", err);
  logError("Uncaught exception", err);
  process.exit(1);
});

process.on("unhandledRejection", (reason) => {
  console.error("Unhandled rejection:", reason);
  logError("Unhandled rejection", reason);
  process.exit(1);
});

try {
  bootstrapKernel({}).then(() => {
    logInfo("Kernel started");
  });
} catch (error) {
  console.error("Failed to bootstrap kernel:", error);
  logError("Failed to bootstrap kernel", error);
  process.exit(1);
}
