export type RepoType = "block" | "virtual";

export interface RepoInitDto {
    name: string
    path: string;
    repoType: RepoType;
    sizeInMb?: number;
}