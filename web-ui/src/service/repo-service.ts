import {ResponseDto} from "@/@types/response-dto.ts";
import {RepoResponseDto} from "@/@types/repo/repo-dto.ts";
import {BACKEND_API_URL} from "@/util/constants.ts";
import {BlockStorageDto} from "@/@types/repo/block-storage-dto.ts";
import {RepoInitDto} from "@/@types/repo/repo-init-dto.ts";
import {BranchNameDto} from "@/@types/repo/branch-name-dto.ts";

export const listRepos = async (): Promise<ResponseDto<RepoResponseDto[]>> => {
    return fetch(BACKEND_API_URL + "/api/repos").then(response => response.json());
};

export const listBlockStorages = async (): Promise<ResponseDto<BlockStorageDto>> => {
    return fetch(BACKEND_API_URL + "/api/repos/block-storages").then(response => response.json());
}

export const listRepoNames = async (): Promise<ResponseDto<BranchNameDto>> => {
    return fetch(BACKEND_API_URL + "/api/repos/names").then(response => response.json());
}

export const initRepo = async (repoInitDto: RepoInitDto): Promise<ResponseDto<RepoResponseDto>> => {
    return fetch(BACKEND_API_URL + "/api/repos", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(repoInitDto)
    }).then(response => response.json());
}