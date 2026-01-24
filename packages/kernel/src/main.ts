import { bootstrapKernel } from "./bootstrap";

try {
  bootstrapKernel({});
} catch (error) {
  console.error("Failed to bootstrap kernel:", error);
  process.exit(1);
}
