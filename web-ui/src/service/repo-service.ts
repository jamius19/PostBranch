import {ResponseDto} from "@/@types/response-dto.ts";
import {RepoResponseDto} from "@/@types/repo-dto.ts";
import {BACKEND_API_URL} from "@/util/constants.ts";
import {BlockStorageDto} from "@/@types/block-storage-dto.ts";

export const listRepos = async (): Promise<ResponseDto<RepoResponseDto[]>> => {
    return fetch(BACKEND_API_URL + "/api/repos").then(response => response.json());
};

export const listBlockStorages = async (): Promise<ResponseDto<BlockStorageDto>> => {
    return fetch(BACKEND_API_URL + "/api/repos/block-storages").then(response => response.json());
}