import type { EventBus } from "../events";
import type { HealthService } from "../health";
import type {
  IRepoRepository,
  IStreamRepository,
  IIssueRepository,
  ISessionRepository,
  IStashRepository,
  IOpRepository,
  ICoreSettingsRepository,
  ISessionSegmentRepository,
  IActiveContextRepository,
} from "../repository";


export interface ICommandContext {
  repos: IRepoRepository;
  streams: IStreamRepository;
  issues: IIssueRepository;
  sessions: ISessionRepository;
  stash: IStashRepository;
  ops: IOpRepository;
  health: HealthService;
  coreSettings: ICoreSettingsRepository
  sessionSegments: ISessionSegmentRepository
  activeContext: IActiveContextRepository

  userId: string;
  deviceId: string;
  now(): string;

  events: EventBus
}
