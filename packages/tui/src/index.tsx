import React, { useEffect, useState } from "react";
import { render, Text } from "ink";
import { App } from "./app.js";
import { ensureKernelRunning } from "./kernel/launcher.js";
import { watchKernel } from "./kernel/watchdog.js";

type KernelInfo = {
  baseUrl: string;
  token: string;
};

let stopWatch: (() => void) | null = null;

function Bootstrap({ onKernelReady }: { onKernelReady: (k: KernelInfo) => void }) {
  const [kernel, setKernel] = useState<KernelInfo | null>(null);

  useEffect(() => {
    ensureKernelRunning()
      .then((k) => {
        setKernel(k);
        onKernelReady(k);
      })
      .catch((err) => {
        console.error("Failed to start kernel:", err);
        process.exit(1);
      });
  }, []);

  if (!kernel) {
    return <Text>Starting Crona kernel…</Text>;
  }

  return <App kernel={kernel} />;
}

const { unmount } = render(
  <Bootstrap
    onKernelReady={(kernel) => {
      stopWatch = watchKernel(kernel, (reason) => {
        // 🔴 kernel died → hard shutdown
        unmount();

        // Full terminal reset (like vim)
        process.stdout.write("\x1Bc");

        console.error(
          "\ncrona: kernel disconnected\n" +
          `reason: ${reason}\n`
        );

        process.exit(1);
      });
    }}
  />,
  {
    stdin: process.stdin,
    stdout: process.stdout,
    exitOnCtrlC: true,
  }
);

const clearAndExit = () => {
  stopWatch?.();
  unmount();
  process.stdout.write("\x1Bc");
  process.exit(0);
};

process.on("SIGINT", clearAndExit);
process.on("SIGTERM", clearAndExit);
process.on("exit", () => {
  process.stdout.write("\x1Bc");
});
