import {ResponseDto} from "@/@types/response-dto.ts";
import {Branch} from "@/@types/repo/repo-response-dto.ts";
import AxiosInstance from "@/service/axios-service.ts";
import {BranchInitDto} from "@/@types/repo/branch-init-dto.ts";

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
