import {PgHostInitDto} from "@/@types/repo/pg/pg-host-init-dto.ts";

export type PgAdapters = PgHostInitDto;
export type PgAdapterName = "host";

export type PgResponseDto = {
    pgConfig: PgAdapters;
    clusterSizeInMb: number;
};
