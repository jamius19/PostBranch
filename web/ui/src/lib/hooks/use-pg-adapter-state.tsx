import {useParams} from "react-router-dom";
import {PgAdapterName, PgAdapters, PgResponseDto} from "@/@types/repo/pg/pg-response-dto.ts";
import {Button} from "@/components/ui/button.tsx";
import {ArrowRight, Check} from "lucide-react";
import Link from "@/components/link.tsx";
import React, {JSX, SyntheticEvent} from "react";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import {reimport} from "@/service/repo-service.ts";
import Spinner from "@/components/spinner.tsx";

type UsePgAdapterStateReturnType = [({pgResponse}: { pgResponse: PgResponseDto }) => JSX.Element]

const usePgAdapterState = (adapter: PgAdapterName): UsePgAdapterStateReturnType => {
    // Undefined repoName means that the repo is not created yet, and it's not re-import workflow
    const {repoName} = useParams<{ repoName: string }>();

    const repoReimport = useNotifiableMutation({
        mutationKey: ["pg-adapter-reimport"],
        mutationFn: (pgConfig: PgAdapters) => reimport(pgConfig, repoName!, adapter),
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
                {!repoName ? (
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
                            <Link to={`/repo/${repoName}`}>
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
