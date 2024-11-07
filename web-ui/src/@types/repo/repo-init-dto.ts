import {RepoPgResponseDto} from "@/@types/repo/repo-pg-init-dto.ts";

export type RepoType = "block" | "virtual";

export interface RepoInitDto {
    name: string
    path: string;
    repoType: RepoType;
    sizeInMb?: number;
}

export interface RepoInitWithPgConfigDto extends RepoInitDto, RepoPgResponseDto {
}
