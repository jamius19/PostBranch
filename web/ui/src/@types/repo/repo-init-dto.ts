import {PgAdapters} from "@/@types/repo/pg/pg-response-dto.ts";

export type RepoType = "block" | "virtual";

export interface RepoInitDto {
    name: string
    path: string;
    repoType: RepoType;
    sizeInMb?: number;
}

export type RepoPgInitDto = {
    repoConfig: RepoInitDto;
    pgConfig: PgAdapters;
}
