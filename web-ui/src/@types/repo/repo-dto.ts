import {RepoType} from "@/@types/repo/repo-init-dto.ts";

export type PgStatus = "COMPLETED" | "STARTED" | "FAILED";

interface Pg {
    id: number;
    version: number;
    status: PgStatus;
    output: string;
}

interface Pool {
    id: number;
    Type: RepoType;
    sizeInMb: number;
    Path: string;
}

interface Branch {
    id: number;
    name: string;
    parentId: number;
    createdAt: Date;
}

export interface RepoResponseDto {
    id: number;
    name: string;
    pg: Pg;
    pool: Pool;
    branches: Branch[];
    createdAt: Date;
    updatedAt: Date;
}
