export type PgStatus = "COMPLETED" | "STARTED" | "FAILED";

interface PgInfo {
    id: number;
    version: number;
    status: PgStatus;
    output: string;
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
    path: string;
    repoType: string;
    sizeInMb: number;
    pg: PgInfo;
    branches: Branch[];
    createdAt: Date;
    updatedAt: Date;
}
