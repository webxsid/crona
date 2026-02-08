import React, { useEffect, useState } from "react";
import { Box, Text, useApp, useInput, useStdout } from "ink";
import { StatusView } from "./views/status.js";

export interface AppProps {
  kernel: {
    baseUrl: string;
    token: string;
  };
}

export function App({ kernel }: AppProps) {
  const { exit } = useApp();
  const { stdout } = useStdout()

  const [size, setSize] = useState<{
    width: number;
    height: number;
  }>({
    width: stdout.columns ?? 80,
    height: stdout.rows ?? 24,
  });

  const { width, height } = size;

  const clearAndExit = () => {
    exit();
  }

  // Handle terminal resize
  useEffect(() => {
    function onResize() {
      setSize({
        width: stdout.columns ?? 80,
        height: stdout.rows ?? 24,
      });
    }

    stdout.on("resize", onResize);
    return () => {
      stdout.off("resize", onResize);
    };
  }, [stdout]);

  useInput((input) => {
    if (input === "q") clearAndExit();
  });


  useEffect(() => {
    console.log("Crona TUI started. Press 'q' to quit.");
    return () => {
      console.log("Crona TUI exited.");
    }
  }, [])

  return (
    <Box
      flexDirection="column"
      width={width}
      height={height}
      padding={1}
    >
      {/* Header */}
      <Box marginBottom={1}>
        <StatusView kernel={kernel} />
      </Box>

      {/* Main body (fills remaining space) */}
      <Box flexGrow={1} borderStyle="round" borderColor="gray">
        {/* future panes go here */}
      </Box>

      {/* Footer */}
      <Box marginTop={1}>
        <Box>
          <Text dimColor>Press q to quit</Text>
        </Box>
      </Box>
    </Box>
  );
}
