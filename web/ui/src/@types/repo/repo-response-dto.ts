import {RepoType} from "@/@types/repo/repo-init-dto.ts";

export type RepoStatus = "READY" | "STARTED" | "FAILED";
export type BranchStatus = "OPEN" | "MERGED" | "CLOSED";
export type BranchPgStatus = "STARTING" | "RUNNING" | "STOPPED" | "FAILED";

export interface Pool {
    id: number;
    Type: RepoType;
    sizeInMb: number;
    path: string;
    mountPath: string;
}

export interface Branch {
    id: number;
    name: string;
    status: BranchStatus;
    pgStatus: BranchPgStatus;
    port: number;
    parentId: number;
    createdAt: Date;
    updatedAt: Date;
}

export interface RepoResponseDto {
    id: number;
    name: string;
    pgVersion: number;
    status: RepoStatus;
    output: string;
    pool: Pool;
    branches: Branch[];
    createdAt: Date;
    updatedAt: Date;
}
