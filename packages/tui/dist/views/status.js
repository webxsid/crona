import React from "react";
import { Box, Text } from "ink";
export function StatusView() {
    return (React.createElement(Box, { borderStyle: "round", borderColor: "cyan", paddingX: 1 },
        React.createElement(Text, null,
            "\u23F1 ",
            React.createElement(Text, { color: "green" }, "IDLE"),
            "  |  No active context")));
}
