import type { ICommandContext } from "./commands";
import type { EventBus } from "./events";
import { HealthService } from "./health";
import {
  ActiveContextRepository,
  CoreSettingsRepository,
  SessionRepository,
  SessionSegmentRepository,
  SqliteIssueRepository,
  SqliteOpRepository,
  SqliteRepoRepository,
  SqliteStreamRepository,
  StashRepository
} from "./repository";
import { dbPing } from "./storage";

export async function createCommandContext(input: {
  userId: string;
  deviceId: string;
  now: () => string;
  events: EventBus;
}): Promise<ICommandContext & { authToken: string }> {

  return {
    userId: input.userId,
    deviceId: input.deviceId,
    now: input.now,

    repos: new SqliteRepoRepository(),
    issues: new SqliteIssueRepository(),
    sessions: new SessionRepository(),
    stash: new StashRepository(),
    ops: new SqliteOpRepository(),
    streams: new SqliteStreamRepository(),
    health: new HealthService({
      dbPing: dbPing,
    }),
    coreSettings: new CoreSettingsRepository(),
    sessionSegments: new SessionSegmentRepository(),
    activeContext: new ActiveContextRepository(),

    events: input.events,

    authToken: crypto.randomUUID(),
  };
}
