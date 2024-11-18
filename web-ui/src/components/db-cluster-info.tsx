import {Info} from "lucide-react";
import {formatValue} from "@/util/lib.ts";
import React from "react";

const DbClusterInfo = ({clusterSizeInMb}: { clusterSizeInMb: number }) => {
    return (
        <>
            <div
                className={"flex items-center gap-3 bg-blue-200/70 text-xs text-blue-700 rounded-md px-4 py-3"}>
                <Info size={16} className={"relative bottom-[1px] flex-shrink-0"}/>

                <p>
                    A superuser named <code className={"font-bold"}>postbranch</code> will be created in
                    the imported database cluster. <br/>
                    <b>Please DO NOT alter<span className={"mx-[1.5px]"}>/</span>delete this user.</b> This user will be
                    used
                    by PostBranch for management
                    purposes.
                </p>
            </div>

            <div
                className={"flex items-center gap-3 bg-lime-200/70 text-xs text-lime-900 rounded-md px-4 py-3"}>
                <Info size={16} className={"relative bottom-[1px] flex-shrink-0"}/>

                <p>
                    Connection successful! The database cluster size
                    is {formatValue(clusterSizeInMb)}.
                </p>
            </div>
        </>
    );
};

export default DbClusterInfo;
