import React from "react";
import { Box, Text } from "ink";

export interface StatusViewProps {
  kernel: {
    baseUrl: string;
    token: string;
  };
}

export function StatusView({ kernel }: StatusViewProps) {
  return (
    <Box
      borderStyle="round"
      borderColor="cyan"
      paddingX={1}
      paddingY={0}
      width="100%"
    >
      <Text>
        <Text bold>Timer:</Text>{" "}
        <Text color="green">IDLE</Text>
        {"  |  "}
        <Text dimColor>No active context</Text>
      </Text>
    </Box>
  );
}
