"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.HealthService = void 0;
class HealthService {
    deps;
    constructor(deps) {
        this.deps = deps;
    }
    async check() {
        return {
            db: await this.deps.dbPing(),
        };
    }
}
exports.HealthService = HealthService;
