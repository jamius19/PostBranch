export type SizeUnit = "K" | "M" | "G";
export type RepoType = "block" | "virtual";

export interface RepoInitDto {
    name: string
    branchName: string;
    path: string;
    repoType: RepoType;
    size?: number;
    sizeUnit?: SizeUnit;
}