import {RepoResponseDto} from "@/@types/repo/repo-response-dto.ts";
import {Activity, Box, Database, GitBranch} from "lucide-react";
import {twMerge as tm} from "tailwind-merge";
import {Link} from "react-router-dom";
import {formatValue} from "@/util/lib.ts";
import React from "react";

interface RepoInfoCardProps {
    repo: RepoResponseDto;
}

const RepoInfoCard = ({repo}: RepoInfoCardProps) => {
    let bgClassNames;

    if (repo.status === 'FAILED') {
        bgClassNames = "border-2 border-red-600 border-dashed text-red-600 shadow-red-500/20 hover:shadow-red-500/20"
    } else if (repo.status === 'READY') {
        bgClassNames = "bg-gray-700 hover:bg-gray-800"
    } else {
        bgClassNames = "bg-gray-500 hover:bg-gray-600"
    }

    return (
        <Link to={`/repo/${repo.name}`}>
            <div
                className={tm("cursor-pointer w-full h-full text-white shadow-lg hover:shadow-xl transition-all duration-500 rounded-lg p-6", bgClassNames)}>
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
                    <span className={"ml-2"}>Postgres {repo.pgVersion}</span>
                </div>

                <div className={"mt-1 text-[0.85rem]"}>
                    <Box size={15} className={"inline-block relative top-[-1.5px]"}/>
                    <span className={"ml-2"}>{formatValue(repo.pool.sizeInMb)}</span>
                </div>

                <div className={"mt-1 text-[0.85rem] font-bold"}>
                    <Activity size={15} className={"inline-block relative top-[-1.5px]"}/>
                    {repo.status === "STARTED" && (
                        <span className={"ml-2"}>Postgres data import in progress</span>
                    )}

                    {repo.status === "READY" && (
                        <span className={"ml-2"}>Repository is ready</span>
                    )}

                    {repo.status === "FAILED" && (
                        <span className={"ml-2"}>Postgres data importing failed</span>
                    )}
                </div>
            </div>
        </Link>
    );
};

export default RepoInfoCard;
