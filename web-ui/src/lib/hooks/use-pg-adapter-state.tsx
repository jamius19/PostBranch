import {useParams} from "react-router-dom";
import {PgAdapterName, PgAdapters, PgResponseDto} from "@/@types/repo/pg/pg-response-dto.ts";
import {Button} from "@/components/ui/button.tsx";
import {ArrowRight, Check} from "lucide-react";
import Link from "@/components/Link.tsx";
import React, {JSX, SyntheticEvent} from "react";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import {reimport} from "@/service/repo-service.ts";
import Spinner from "@/components/Spinner.tsx";

type UsePgAdapterStateReturnType = [({pgResponse}: { pgResponse: PgResponseDto }) => JSX.Element]

const usePgAdapterState = (adapter: PgAdapterName): UsePgAdapterStateReturnType => {
    const {repoId} = useParams<{ repoId?: string }>();
    // -1 means that the repo is not created yet, and it's not re-import workflow
    const repoIdInt = parseInt(repoId ?? "-1");

    const repoReimport = useNotifiableMutation({
        mutationKey: ["pg-reimport"],
        mutationFn: (pgConfig: PgAdapters) => reimport(pgConfig, repoIdInt, adapter),
        messages: {
            pending: "Starting Postgres import",
            success: "Postgres import started successfully",
        },
        invalidates: ["repo-list", "repo"],
    });

    const handleImport = async (e: SyntheticEvent, pgConfig: PgAdapters) => {
        e.preventDefault()
        await repoReimport.mutateAsync(pgConfig);
    }

    const Nav = ({pgResponse}: { pgResponse: PgResponseDto }) => {
        return (
            <div className={"flex gap-4"}>
                {repoIdInt === -1 ? (
                    <Link to={"/repo/setup/storage"}
                          state={{pgResponse, adapter}}>
                        <Button>
                            Configure Storage <ArrowRight/>
                        </Button>
                    </Link>
                ) : (
                    <>
                        <Button
                            onClick={(e) => handleImport(e, pgResponse.pgConfig)}
                            disabled={repoReimport.isSuccess || repoReimport.isPending}
                            variant={repoReimport.isSuccess ? "success" : "default"}>

                            <Spinner isLoading={repoReimport.isPending} light/>
                            {repoReimport.isSuccess && <Check/>}
                            {repoReimport.isSuccess ? "Import Started" : "Start Import"}
                        </Button>

                        {repoReimport.isSuccess && (
                            <Link to={`/repo/${repoIdInt}`}>
                                <Button>
                                    Go to Repository <ArrowRight/>
                                </Button>
                            </Link>
                        )}
                    </>
                )}
            </div>
        );
    }

    return [Nav];
};

export default usePgAdapterState;
