import {Copy} from "lucide-react";
import copy from "copy-to-clipboard";
import {toast} from "react-toastify";
import React from "react";
import {cn} from "@/lib/utils.ts";

const CopyToClipboard = (
    {data, className}: { data: string, className?: string }
) => {
    return (
        <Copy className={cn("relative bottom-[1.5px] hover:text-blue-500 cursor-pointer", className)}
              size={15}
              onClick={() => {
                  copy(data);
                  toast.success("Copied to clipboard");
              }}/>
    );
};

export default CopyToClipboard;
