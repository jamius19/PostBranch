import {RepoResponseDto} from "@/@types/repo/repo-dto.ts";
import {Activity, Database, GitBranch} from "lucide-react";
import {twMerge as tm} from "tailwind-merge";

interface RepoInfoCardProps {
    repo: RepoResponseDto;
}

const RepoInfoCard = ({repo}: RepoInfoCardProps) => {
    let bgClassNames;

    if (repo.pg) {
        if (repo.pg.status === 'FAILED') {
            bgClassNames = "border-2 border-red-600 border-dashed text-red-600 shadow-red-500/20 hover:shadow-red-500/20"
        } else {
            bgClassNames = "bg-gray-700 hover:bg-gray-800"
        }
    } else {
        bgClassNames = "bg-gray-500 hover:bg-gray-600"
    }

    return (
        <div
            className={tm("cursor-pointer w-full h-full text-white hover:shadow-xl transition-all duration-500 rounded-lg shadow-lg p-6", bgClassNames)}>
            <h3 className="font-title text mono font-bold">{repo.name}</h3>

            <div className={"mt-3 text-[0.85rem]"}>
                <GitBranch size={15} className={"inline-block relative top-[-1.5px]"}/>
                <span className={"ml-2"}>
                        {repo.branches.length ? `${repo.branches.length}` : "No"}&nbsp;
                    {repo.branches.length > 1 ? "Branches" : "Branch"}
                    </span>
            </div>
            <div className={"mt-1 text-[0.85rem]"}>
                <Database size={15} className={"inline-block relative top-[-1.5px]"}/>
                {repo.pg ? (
                    <span className={"ml-2"}>Postgres {repo.pg.version}</span>
                ) : (
                    <span className={"ml-2"}>Postgres not imported</span>
                )}
            </div>

            <div className={"mt-1 text-[0.85rem] font-bold"}>
                <Activity size={15} className={"inline-block relative top-[-1.5px]"}/>
                {repo.pg ? (
                    <>
                        {repo.pg.status === "STARTED" && (
                            <span className={"ml-2"}>Postgres data import in progress</span>
                        )}

                        {repo.pg.status === "COMPLETED" && (
                            <span className={"ml-2"}>Repository is ready</span>
                        )}

                        {repo.pg.status === "FAILED" && (
                            <span className={"ml-2"}>Postgres data importing failed</span>
                        )}
                    </>
                ) : (
                    <span className={"ml-2"}>Postgres not imported</span>
                )}
            </div>
        </div>
    );
};

export default RepoInfoCard;
