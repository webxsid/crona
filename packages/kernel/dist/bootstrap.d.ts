export interface IBootstrapKernelOptions {
    dbPath?: string | undefined;
}
export declare function bootstrapKernel({ dbPath }: IBootstrapKernelOptions): Promise<void>;
