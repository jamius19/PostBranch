import {PgLocalInitDto} from "@/@types/repo/pg/pg-local-init-dto.ts";
import {PgHostInitDto} from "@/@types/repo/pg/pg-host-init-dto.ts";

export type PgAdapters = PgLocalInitDto | PgHostInitDto;
export type PgAdapterNames = "local" | "host";

export type PgResponseDto = {
    pgConfig: PgAdapters;
    clusterSizeInMb: number;
};
