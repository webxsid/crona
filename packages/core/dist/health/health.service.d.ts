export interface IHealthStatus {
    db: boolean;
}
export declare class HealthService {
    private readonly deps;
    constructor(deps: {
        dbPing: () => Promise<boolean>;
    });
    check(): Promise<IHealthStatus>;
}
