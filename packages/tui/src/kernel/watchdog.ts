// kernel/watchdog.ts
export function watchKernel(
  kernel: { baseUrl: string },
  onDeath: (reason: string) => void
) {
  const interval = setInterval(async () => {
    try {
      const res = await fetch(`${kernel.baseUrl}/health`);
      if (!res.ok) throw new Error("Health check failed");
    } catch {
      clearInterval(interval);
      onDeath("Kernel process exited");
    }
  }, 1000);

  return () => clearInterval(interval);
}
