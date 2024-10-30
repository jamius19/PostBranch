import {ResponseDto} from "@/@types/response-dto.ts";
import {RepoResponseDto} from "@/@types/repo/repo-dto.ts";
import {BlockStorageDto} from "@/@types/repo/block-storage-dto.ts";
import {RepoInitDto} from "@/@types/repo/repo-init-dto.ts";
import AxiosInstance from "@/service/axios-service.ts";
import {RepoPgInitDto} from "@/@types/repo/repo-pg-init-dto.ts";

export const listRepos = async (): Promise<ResponseDto<RepoResponseDto[]>> => {
    return AxiosInstance.get("/api/repos");
};

export const listBlockStorages = async (): Promise<ResponseDto<BlockStorageDto>> => {
    return AxiosInstance.get("/api/repos/block-storages");
}

export const initRepo = async (repoInitDto: RepoInitDto): Promise<ResponseDto<RepoResponseDto>> => {
    return AxiosInstance.post("/api/repos", repoInitDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}

export const importPg = async (data: {
    repoId: number,
    repoPgInitDto: RepoPgInitDto
}): Promise<ResponseDto<RepoResponseDto>> => {
    return AxiosInstance.post(`/api/repos/${data.repoId}/postgres`, data.repoPgInitDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}