export interface ActiveContext {
  userId: string;
  deviceId: string;
  repoId?: string | undefined;
  streamId?: string | undefined;
  issueId?: string | undefined;
  updatedAt: Date;
}
