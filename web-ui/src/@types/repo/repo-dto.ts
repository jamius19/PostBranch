interface PgInfo {
    id: number;
    version: number;
    status: string;
    output: string;
}

interface Branch {
    id: number;
    name: string;
    parentId: number;
}

export interface RepoResponseDto {
    id: number;
    name: string;
    path: string;
    repoType: string;
    size: number;
    sizeUnit: string;
    pg?: PgInfo;
    branches: Branch[];
    createdAt: Date;
    updatedAt: Date;
}
