import {Info} from "lucide-react";
import {formatValue} from "@/util/lib.ts";
import React from "react";

const DbClusterInfo = ({clusterSizeInMb}: { clusterSizeInMb: number }) => {
    return (
        <div
            className={"flex items-center gap-3 bg-lime-200/70 text-xs text-lime-900 rounded-md px-4 py-3"}>
            <Info size={16} className={"relative bottom-[1px] flex-shrink-0"}/>

            <p>
                Connection successful! The database cluster size is {formatValue(clusterSizeInMb)}.
            </p>
        </div>
    );
};

export default DbClusterInfo;
