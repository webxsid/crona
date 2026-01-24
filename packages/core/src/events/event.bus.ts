import type { KernelEvent } from "./event.types";

type Listener = (event: KernelEvent) => void;

export class EventBus {
  private listeners = new Set<Listener>();

  emit(event: KernelEvent) {
    for (const listener of this.listeners) {
      listener(event);
    }
  }

  subscribe(listener: Listener) {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }
}
