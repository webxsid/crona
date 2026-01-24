import type { Op } from "../../domain";

export interface IOpRepository {
  append(op: Op): Promise<void>;

  latest(limit: number): Promise<Op[]>;

  listSince(
    userId: string,
    sinceTimestamp: string
  ): Promise<Op[]>;

  listByEntity(
    entity: Op["entity"],
    entityId: string,
    userId: string,
    limit: number | undefined
  ): Promise<Op[]>;
}
