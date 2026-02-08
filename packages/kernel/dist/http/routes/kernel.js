"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.KernelRoutes = void 0;
const kernel_info_1 = require("../../kernel-info");
class KernelRoutes {
    app;
    _ctx;
    constructor(app, _ctx) {
        this.app = app;
        this._ctx = _ctx;
    }
    register() {
        this.app.get("/kernel/info", async () => {
            const info = await (0, kernel_info_1.readKernelInfo)();
            return info;
        });
    }
}
exports.KernelRoutes = KernelRoutes;
