export type StreamVisibility = "personal" | "shared";

export interface Stream {
  id: string;
  repoId: string;
  name: string;
  visibility: StreamVisibility;
}
