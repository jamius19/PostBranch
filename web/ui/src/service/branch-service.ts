import {ResponseDto} from "@/@types/response-dto.ts";
import {Branch} from "@/@types/repo/repo-response-dto.ts";
import AxiosInstance from "@/service/axios-service.ts";
import {BranchInitDto} from "@/@types/repo/branch-init-dto.ts";
import {BranchCloseDto} from "@/@types/repo/branch-close-dto";

export const initBranch = async (
    repoName: string,
    branchInit: BranchInitDto,
): Promise<ResponseDto<Branch>> => {

    return AxiosInstance.post(`/api/repos/${repoName}/branch`, branchInit, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}

export const closeBranch = async (
    repoName: string,
    branchCloseDto: BranchCloseDto,
): Promise<ResponseDto<number>> => {

    return AxiosInstance.post(`/api/repos/${repoName}/branch/close`, branchCloseDto, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}
