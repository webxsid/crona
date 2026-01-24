import type { KernelEvent } from "./event.types";
type Listener = (event: KernelEvent) => void;
export declare class EventBus {
    private listeners;
    emit(event: KernelEvent): void;
    subscribe(listener: Listener): () => boolean;
}
export {};
