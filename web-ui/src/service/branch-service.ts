import {ResponseDto} from "@/@types/response-dto.ts";
import {Branch} from "@/@types/repo/repo-dto.ts";
import AxiosInstance from "@/service/axios-service.ts";
import {BranchInitDto} from "@/@types/repo/branch-init-dto.ts";

export const initBranch = async (
    repoId: number,
    branchInit: BranchInitDto,
): Promise<ResponseDto<Branch>> => {

    return AxiosInstance.post(`/api/repos/${repoId}/branch`, branchInit, {
        headers: {
            "Content-Type": "application/json"
        }
    });
}