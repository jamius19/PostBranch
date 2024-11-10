import {Navigate, useNavigate, useParams} from "react-router-dom";
import React from "react";
import {formatValue, isInteger} from "@/util/lib.ts";
import {useQuery} from "@tanstack/react-query";
import {deleteRepo, getRepo} from "@/service/repo-service.ts";
import Spinner from "@/components/Spinner.tsx";
import {ArrowRight, Box, CircleCheck, Clock2, Database, GitBranch, OctagonX, Trash2, TriangleAlert} from "lucide-react";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table.tsx";
import dayjs from "dayjs";
import utc from 'dayjs/plugin/utc'
import {PgStatus} from "@/@types/repo/repo-dto.ts";
import {Button} from "@/components/ui/button.tsx";
import {clsx} from "clsx";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import Link from "@/components/Link.tsx";

const Repo = () => {
    const navigate = useNavigate();

    dayjs.extend(utc);
    const {repoId: repoIdStr} = useParams<{ repoId: string }>();
    const repoId = parseInt(repoIdStr!);

    const repoQuery = useQuery({
        queryKey: ["repo", repoId],
        queryFn: () => getRepo(repoId),
    });

    const repoDeleteQuery = useNotifiableMutation({
        mutationKey: ["repo-delete", repoId],
        mutationFn: () => deleteRepo(repoId),
        invalidates: ["repo-list"],
        messages: {
            pending: "Deleting repository",
            success: "Repository deleted successfully."
        },
        onSuccess: () => {
            navigate("/");
        }
    });

    const handleDelete = () => {
        repoDeleteQuery.mutate(repoId);
    }

    const repo = repoQuery.data?.data;

    const disableInteraction = repoDeleteQuery.isPending;

    if (!isInteger(repoIdStr)) {
        return <Navigate to={"/error"} state={{message: "The repository ID in the URL is invalid."}}/>;
    }

    if (repoQuery.isPending || disableInteraction || repoDeleteQuery.isSuccess || repoQuery.isRefetching) {
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

            <InfoBlock status={repo.pg.status}/>

            <div className={"flex mt-2 mb-12 flex-col gap-2 text-sm"}>
                <p>
                    <Box
                        className={"inline-block relative top-[-1px] me-1.5"}
                        size={16}/>
                    {formatValue(repo.pool.sizeInMb)}
                </p>

                <p>
                    <Database
                        className={"inline-block relative top-[-1.5px] me-1.5"}
                        size={15}/>
                    Postgres {repo.pg.version}
                </p>
            </div>

            {repo.pg.status === "COMPLETED" && (
                <div className={""}>
                    <div className={"mb-2 flex items-center gap-3"}>
                        <GitBranch size={22} className={"relative top-[-5.5px]"}/>
                        <h1>Branches</h1>
                    </div>

                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[100px]">Name</TableHead>
                                <TableHead>Branched From</TableHead>
                                <TableHead className="text-right">Created</TableHead>
                            </TableRow>
                        </TableHeader>

                        <TableBody>
                            {repo.branches.length !== 0 ? repo.branches.map((branch) => (
                                <TableRow key={branch.id}>
                                    <TableCell className="font-medium w-[300px]">{branch.name}</TableCell>
                                    <TableCell className={"w-[300px]"}>{branch.parentId ?? "—"}</TableCell>
                                    <TableCell className="text-right">
                                        {dayjs.utc(repo.createdAt).format("DD/MM/YYYY HH:mm:ss")}
                                    </TableCell>
                                </TableRow>
                            )) : (
                                <TableRow>
                                    <TableCell colSpan={3} className={"text-center text-muted-foreground/80"}>
                                        No branches found
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </div>
            )}

            {repo.pg.status === "FAILED" && (
                <div>
                    <div>
                        Output of the last import attempt. Please consult the PostBranch log for more details.
                    </div>

                    <div
                        className="mt-3  min-h-[300px] max-h-[800px] overflow-x-clip overflow-y-auto border border-gray-400/50 rounded-md p-4 mono text-xs flex flex-col gap-1">
                        {repo.pg.output.split(";").map((line, index) => {
                            const mainError = !line.match(/.*<nil>$/);

                            return <p key={index} className={clsx(mainError && "text-red-700 font-bold")}>{line}</p>;
                        })}
                    </div>

                    <Link
                        disabled={disableInteraction}
                        to={`/repo/setup/${repoId}/postgres`}
                        className={"mt-4 block"}>
                        <Button disabled={disableInteraction}>
                            Change Postgres Config <ArrowRight/>
                        </Button>
                    </Link>

                    <div className={"mt-8"}>
                        <Link
                            disabled={disableInteraction}
                            to={`/repo/setup/${repoId}/postgres`}
                            className={"mt-4 block"}>
                            <Button disabled={disableInteraction}>
                                Import Postgres <ArrowRight/>
                            </Button>
                        </Link>
                    </div>
                </div>
            )}
        </div>
    );
};

const InfoBlock = ({status}: { status?: PgStatus }) => {
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


    if (status === "COMPLETED") {
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

export default Repo;
