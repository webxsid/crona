import type { Stream, StreamVisibility } from "../../domain";

export interface IStreamRepository {
  create(
    stream: Stream,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<Stream>;

  getById(
    streamId: string,
    userId: string
  ): Promise<Stream | null>;

  listByRepo(
    repoId: string,
    userId: string
  ): Promise<Stream[]>;

  update(
    streamId: string,
    updates: {
      name?: string | undefined;
      visibility?: StreamVisibility | undefined;
    },
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<Stream>;

  softDelete(
    streamId: string,
    meta: {
      userId: string;
      now: string;
    }
  ): Promise<void>;
}
