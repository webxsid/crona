"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const bootstrap_1 = require("./bootstrap");
try {
    (0, bootstrap_1.bootstrapKernel)({});
}
catch (error) {
    console.error("Failed to bootstrap kernel:", error);
    process.exit(1);
}
