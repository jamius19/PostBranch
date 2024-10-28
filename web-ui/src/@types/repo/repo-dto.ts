export interface RepoResponseDto {
    id: number;
    name: string;
    path: string;
    repoType: string;
    size: number;
    sizeUnit: string;
    pg_id?: number;
    pool_id: number;
    created_at: string;
    updated_at: string;
}