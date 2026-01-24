export interface IHealthStatus {
  db: boolean;
}

export class HealthService {
  constructor(private readonly deps: { dbPing: () => Promise<boolean> }) { }

  async check(): Promise<IHealthStatus> {
    return {
      db: await this.deps.dbPing(),
    };
  }
}
