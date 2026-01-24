import React from "react";
import { Box } from "ink";
import { StatusView } from "./views/status.js";
export function App() {
    return (React.createElement(Box, { flexDirection: "column", padding: 1 },
        React.createElement(StatusView, null)));
}
