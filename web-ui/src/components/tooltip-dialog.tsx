import {Dialog, DialogContent, DialogTrigger} from "@/components/ui/dialog.tsx";
import {Tooltip, TooltipContent, TooltipProvider, TooltipTrigger,} from "@/components/ui/tooltip";
import React, {JSX, useState} from "react";
import {cn} from "@/lib/utils.ts";

type TooltipDialogProps = {
    icon: JSX.Element;
    tooltip: React.ReactNode;
    children: React.ReactNode;
    className?: string;
    open?: boolean;
    showClose?: boolean;
};

const TooltipDialog = (props: TooltipDialogProps) => {
    const [open, setopen] = useState(false);

    return (
        <TooltipProvider>
            <Dialog
                modal={true}
                onOpenChange={props.open ? undefined : setopen}
                open={props.open ?? open}>

                <Tooltip>
                    <TooltipTrigger asChild>
                        <DialogTrigger asChild>
                            <div
                                className={cn("border border-gray-300 hover:border-gray-800 hover:bg-gray-800 transition-all duration-100 rounded px-1 py-1 text-gray-600 hover:text-white relative bottom-[1.5px] cursor-pointer", props.className)}>
                                {props.icon}
                            </div>
                        </DialogTrigger>
                    </TooltipTrigger>
                    <TooltipContent>
                        {props.tooltip}
                    </TooltipContent>
                </Tooltip>

                <DialogContent
                    className={"w-[1000px]"}
                    showClose={props.showClose}>
                    {props.children}
                </DialogContent>
            </Dialog>
        </TooltipProvider>
    );
};

export default TooltipDialog;
