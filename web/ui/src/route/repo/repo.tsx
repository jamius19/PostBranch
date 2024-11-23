import {Navigate, useNavigate, useParams} from "react-router-dom";
import React, {JSX, useMemo} from "react";
import {formatValue} from "@/util/lib.ts";
import {useQuery} from "@tanstack/react-query";
import {deleteRepo, getRepo} from "@/service/repo-service.ts";
import Spinner from "@/components/spinner.tsx";
import {
    ArrowRight,
    Box,
    CircleCheck,
    Clock2,
    Database,
    GitBranch,
    HardDrive,
    Info,
    OctagonX,
    Trash2,
    TriangleAlert,
    Wrench,
    X
} from "lucide-react";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table.tsx";
import dayjs from "dayjs";
import utc from 'dayjs/plugin/utc'
import {Branch, BranchPgStatus, RepoStatus, RepoResponseDto} from "@/@types/repo/repo-response-dto.ts";
import {Button} from "@/components/ui/button.tsx";
import {clsx} from "clsx";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import Link from "@/components/link.tsx";
import {twMerge as tm} from "tailwind-merge";
import {DialogFooter, DialogHeader, DialogTitle,} from "@/components/ui/dialog.tsx"
import {Badge} from "@/components/ui/badge.tsx";
import CopyToClipboard from "@/components/copy-to-clipboard.tsx";
import TooltipDialog from "@/components/tooltip-dialog.tsx";
import {Tooltip, TooltipContent, TooltipProvider, TooltipTrigger} from "@/components/ui/tooltip.tsx";
import NewBranch from "@/components/new-branch.tsx";


const Repo = () => {
    const {repoName} = useParams<{ repoName: string }>();

    if (!repoName) {
        throw new Error("Repo name is required");
    }

    const navigate = useNavigate();
    dayjs.extend(utc);

    const repoDeleteQuery = useNotifiableMutation({
        mutationKey: ["repo-delete", repoName],
        mutationFn: deleteRepo,
        invalidates: ["repo-list"],
        messages: {
            pending: "Deleting repository",
            success: "Repository deleted successfully."
        },
        onSuccess: () => {
            navigate("/");
        }
    });

    const repoQuery = useQuery({
        queryKey: ["repo", repoName],
        queryFn: () => getRepo(repoName),
        refetchInterval: 2000,
        enabled: !repoDeleteQuery.isPending && !repoDeleteQuery.isSuccess,
    });

    const handleDelete = () => {
        repoDeleteQuery.mutate(repoName);
    }

    const repo = repoQuery.data?.data;

    const disableInteraction = repoDeleteQuery.isPending;

    const branchMap = useMemo(() => {
        const branchMap = new Map<number, Branch>();

        if (!repo?.branches) {
            return new Map<number, Branch>();
        }

        repo.branches.forEach(branch => {
            branchMap.set(branch.id, branch);
        });

        return branchMap;
    }, [repo]);

    const getParentBranchName = (branchId?: number): string => {
        if (!branchId) {
            return "—";
        }

        const branch = branchMap.get(branchId);

        if (!branch) {
            console.error(`Branch with ID ${branchId} not found.`);
            return "—";
        }

        return branch.name;
    }

    if (repoQuery.isPending || disableInteraction || repoDeleteQuery.isSuccess) {
        return <Spinner/>;
    }

    if (!repoQuery.isSuccess || !repo) {
        return <Navigate to={"/error"} state={{message: "An error occurred while fetching the repository."}}/>;
    }

    return (
        <div>
            <div className={"flex mb-2.5 items-center gap-3 "}>
                <h1 className={"mono"}>{repo.name}</h1>
                <Button
                    className={"relative top-[-3px] ml-auto text-gray-400 px-2 py-2 hover:bg-red-600 hover:text-white hover:border-red-600 hover:shadow-md hover:shadow-red-500/40 transition-all duration-200"}
                    variant={"ghost"}
                    onClick={handleDelete}
                    disabled={disableInteraction}>
                    <Trash2/>
                </Button>
            </div>

            <PgInfoBlock status={repo.status}/>

            <div className={"flex mt-3 mb-20 flex-col gap-2 text-sm"}>
                <div>
                    <Box
                        className={"inline-block relative top-[-1px] me-2.5"}
                        size={16}/>
                    {formatValue(repo.pool.sizeInMb)}
                </div>

                <div>
                    <Database
                        className={"inline-block relative top-[-1.5px] me-2.5"}
                        size={15}/>
                    Postgres {repo.pgVersion}
                </div>

                <div>
                    <HardDrive
                        className={"inline-block relative top-[-1.5px] me-2.5"}
                        size={15}/>
                    <code className={"text-[0.83rem]"}>{repo.pool.mountPath}</code>
                </div>
            </div>

            {repo.status === "READY" && (
                <div className={""}>
                    <div className={"mb-2 flex items-center gap-3"}>
                        <GitBranch size={22} className={"relative top-[-5.5px]"}/>
                        <h1>Branches</h1>

                        <div className={"relative bottom-0.5 ml-auto"}>
                            <NewBranch repoName={repo.name} branches={repo.branches} branchMap={branchMap}/>
                        </div>
                    </div>

                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[100px]">Name</TableHead>
                                <TableHead className="w-[150px]">Actions</TableHead>
                                <TableHead>Port</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className={"w-[100px]"}>Branched From</TableHead>
                            </TableRow>
                        </TableHeader>

                        <TableBody>
                            {repo.branches.length !== 0 ? repo.branches.map((branch) => (
                                <TableRow key={branch.id}>
                                    <TableCell className="font-medium w-[200px] flex">
                                        <span>{branch.name}</span>
                                        {branch.name === "main" && (
                                            <Badge
                                                className={"ml-2 opacity-80 rounded-full"}
                                                variant={"info"}
                                                size={"sm"}>
                                                Primary
                                            </Badge>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <BranchActions repo={repo} branch={branch}/>
                                    </TableCell>
                                    <TableCell className={"w-[100px]"}>{branch.port}</TableCell>
                                    <TableCell className={"w-[150px]"}>
                                        <BranchPgStatusBadge status={branch.pgStatus}/>
                                    </TableCell>
                                    <TableCell
                                        className={"w-[200px]"}>
                                        {getParentBranchName(branch.parentId)}
                                    </TableCell>
                                </TableRow>
                            )) : (
                                <TableRow>
                                    <TableCell colSpan={5} className={"text-center text-muted-foreground/80"}>
                                        No branches found
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            )}

            {repo.status === "FAILED" && (
                <div>
                    <div>
                        Output of the last import attempt. Please consult the PostBranch log for more details.
                    </div>

                    <div
                        className="mt-3  min-h-[300px] max-h-[800px] overflow-x-clip overflow-y-auto border border-gray-400/50 rounded-md p-4 mono text-xs flex flex-col gap-1">
                        {repo.output.split(";").map((line, index) => {
                            const mainError = !line.match(/.*<nil>$/);

                            return <p key={index} className={clsx(mainError && "text-red-700 font-bold")}>{line}</p>;
                        })}
                    </div>

                    <Link
                        disabled={disableInteraction}
                        to={`/repo/setup/${repoName}/postgres`}
                        className={"mt-4 block"}>
                        <Button disabled={disableInteraction}>
                            Change Postgres Config <ArrowRight/>
                        </Button>
                    </Link>
                </div>
            )}
        </div>
    );
};

const PgInfoBlock = ({status}: { status?: RepoStatus }): JSX.Element => {
    if (!status) {
        return (
            <div>
                <p className={"bg-amber-700 text-white inline-block ps-2 pe-3 py-1.5 rounded-md text-xs"}>
                    <TriangleAlert className={"inline-block relative top-[-1px] me-1"} size={14}/>
                    Postgres not imported
                </p>
            </div>
        );
    }


    if (status === "READY") {
        return (
            <div>
                <p className={"bg-lime-600 text-white inline-block ps-2 pe-3 py-1.5 rounded-md text-xs"}>
                    <CircleCheck className={"inline-block relative top-[-1px] me-1"} size={14}/>
                    Repository is ready
                </p>
            </div>
        );
    } else if (status === "STARTED") {
        return (
            <div>
                <p className={"bg-amber-700 text-white inline-block ps-2 pe-3 py-1.5 rounded-md text-xs"}>
                    <Clock2 className={"inline-block relative top-[-1px] me-1"} size={14}/>
                    Postgres import in progress
                </p>
            </div>
        );
    } else {
        return (
            <div>
                <p className={"bg-red-600 text-white inline-block ps-2 pe-3 py-1.5 rounded-md text-xs"}>
                    <OctagonX className={"inline-block relative top-[-1.5px] me-1"} size={14}/>
                    Postgres import failed
                </p>
            </div>
        );
    }
}

const BranchActions = (
    {repo, branch}: { repo: RepoResponseDto, branch: Branch }
): JSX.Element => {

    return (
        <div className={"flex gap-1"}>
            <TooltipDialog
                tooltip={<p>View additional information about <code>{branch.name}</code> branch</p>}
                icon={<Info size={14}/>}>

                <DialogHeader>
                    <DialogTitle>Branch Information</DialogTitle>
                </DialogHeader>

                <div className={"text-sm text-gray-700 flex flex-col gap-5"}>
                    <p>
                        <b>Postgres Port:</b><br/><code>{branch.port}</code>
                        <CopyToClipboard
                            className={"inline ms-2"}
                            data={`${branch.port}`}/>
                    </p>

                    <p>
                        <b>Postgres Data Cluster
                            Path:</b><br/><code>{`${repo.pool.mountPath}/${branch.name}/data`}</code>
                        <CopyToClipboard
                            className={"inline ms-2"}
                            data={`${repo.pool.mountPath}/${branch.name}/data`}/>
                    </p>

                    <p>
                        <b>Postgres Log Path:</b><br/><code>{`${repo.pool.mountPath}/${branch.name}/logs`}</code>
                        <CopyToClipboard
                            className={"inline ms-2"}
                            data={`${repo.pool.mountPath}/${branch.name}/logs`}/>
                    </p>

                    <p>
                        <b>Created Date:</b><br/> {dayjs.utc(branch.createdAt).format("DD MMM, YYYY HH:mm:ss")} UTC
                    </p>

                    <p>
                        <b>Updated Date:</b><br/> {dayjs.utc(branch.createdAt).format("DD MMM, YYYY HH:mm:ss")} UTC
                    </p>
                </div>
            </TooltipDialog>

            <TooltipProvider>
                <Tooltip>
                    <TooltipTrigger>
                        <Link to={"/repos/" + repo.id + "/branches/" + branch.id}>
                            <div
                                className={"border border-gray-300 hover:border-gray-800 hover:bg-gray-800 transition-all duration-100 rounded px-1 py-1 text-gray-600 hover:text-white relative bottom-[1.5px]"}>
                                <Wrench size={14}/>
                            </div>
                        </Link>
                    </TooltipTrigger>
                    <TooltipContent>
                        Configure Branch Settings
                    </TooltipContent>
                </Tooltip>
            </TooltipProvider>

            <TooltipDialog
                className={"hover:border-red-500 hover:bg-red-500"}
                tooltip={<p>Close <code>{branch.name}</code> branch</p>}
                icon={<X size={14}/>}>

                <DialogHeader>
                    <DialogTitle>Close Branch</DialogTitle>
                </DialogHeader>

                <div className={"text-sm text-gray-700"}>
                    <p>
                        Are you sure you want to close this branch?<br/>
                        The postgres instance will be shut down and associated data will be deleted.
                    </p>

                    <p className={"mt-3 font-bold"}>
                        This action cannot be undone.
                    </p>
                </div>
                <DialogFooter>
                    <Button variant={"ghost"} size={"sm"}>Cancel</Button>
                    <Button variant={"destructive"} size={"sm"}>
                        Close Branch
                    </Button>
                </DialogFooter>
            </TooltipDialog>

        </div>
    );
}

const BranchPgStatusBadge = (
    {status}: { status: BranchPgStatus }
): JSX.Element => {
    const starting = status === "STARTING";
    const running = status === "RUNNING";
    const stopped = status === "STOPPED";
    const failed = status === "FAILED";

    return (
        <div
            className={tm(
                "select-none text-primary inline-flex items-center justify-center rounded-full border ps-2.5 pe-3 text-[11px] font-bold transition-all duration-200",
                running && "border-gray-800 bg-gray-800 text-white",
                failed && "border-red-500 bg-white text-red-500",
            )}>

            {!stopped && !starting ? (
                <span className={tm(
                    "relative top-[-1.5px] text-[23px] me-1",
                    running && "text-lime-500",
                    failed && "text-red-500",
                )}>
                ●
                </span>
            ) : (
                <span className={"relative top-[-1px] text-[23px] me-1 text-gray-500"}>
                ○
                </span>
            )}

            {status}
        </div>
    )
}

export default Repo;
