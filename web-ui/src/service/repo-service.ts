import {ResponseDto} from "@/@types/response-dto.ts";
import {RepoResponseDto} from "@/@types/repo/repo-dto.ts";
import {RepoPgInitDto} from "@/@types/repo/repo-init-dto.ts";
import AxiosInstance from "@/service/axios-service.ts";
import {PgAdapterName, PgAdapters, PgResponseDto} from "@/@types/repo/pg/pg-response-dto.ts";

export const listRepos = async (): Promise<ResponseDto<RepoResponseDto[]>> => {
    return AxiosInstance.get("/api/repos");
};

export const getRepo = async (repoId: number): Promise<ResponseDto<RepoResponseDto>> => {
    return AxiosInstance.get(`/api/repos/${repoId}`);
};

export const initRepo = async (repoPgInitDto: RepoPgInitDto, type: PgAdapterName):
    Promise<ResponseDto<RepoResponseDto>> => {

    return AxiosInstance.post(`/api/repos/import/${type}`, repoPgInitDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}

export const reimport = async (
    pgInitDto: PgAdapters,
    repoId: number,
    type: PgAdapterName,
): Promise<ResponseDto<RepoResponseDto>> => {

    return AxiosInstance.post(`/api/repos/import/${repoId}/${type}`, pgInitDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}

export const validatePg = async <T extends PgAdapters, >
(repoPgInitDto: T, type: PgAdapterName):
    Promise<ResponseDto<PgResponseDto>> => {

    return AxiosInstance.post(`/api/repos/postgres/validate/${type}`, repoPgInitDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}

export const deleteRepo = async (repoId: number): Promise<ResponseDto<number>> => {
    return AxiosInstance.delete(`/api/repos/${repoId}`);
}

