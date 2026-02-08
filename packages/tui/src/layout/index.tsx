import React from "react";
import { Box } from "ink";
import { useStdoutDimensions } from "../hooks/useStdoutDimesions.hook.js";

export function Layout({ children }: { children: React.ReactNode }) {
  const { columns, rows } = useStdoutDimensions();

  return (
    <Box width={columns} height={rows} flexDirection="column">
      {children}
    </Box>
  );
}
