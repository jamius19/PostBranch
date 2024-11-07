import {ResponseDto} from "@/@types/response-dto.ts";
import {RepoResponseDto} from "@/@types/repo/repo-dto.ts";
import {RepoInitWithPgConfigDto} from "@/@types/repo/repo-init-dto.ts";
import AxiosInstance from "@/service/axios-service.ts";
import {RepoPgInitDto, RepoPgResponseDto} from "@/@types/repo/repo-pg-init-dto.ts";

export const listRepos = async (): Promise<ResponseDto<RepoResponseDto[]>> => {
    return AxiosInstance.get("/api/repos");
};

export const getRepo = async (repoId: number): Promise<ResponseDto<RepoResponseDto>> => {
    return AxiosInstance.get(`/api/repos/${repoId}`);
};

export const initRepo = async (repoInitWithPgDto: RepoInitWithPgConfigDto):
    Promise<ResponseDto<RepoResponseDto>> => {

    return AxiosInstance.post("/api/repos", repoInitWithPgDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}

export const validatePg = async (repoPgInitDto: RepoPgInitDto):
    Promise<ResponseDto<RepoPgResponseDto>> => {

    return AxiosInstance.post("/api/postgres", repoPgInitDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}

export const deleteRepo = async (repoId: number): Promise<ResponseDto<number>> => {
    return AxiosInstance.delete(`/api/repos/${repoId}`);
}

