import {useQuery} from "@tanstack/react-query";
import {Navigate} from "react-router-dom";
import {listRepos} from "@/service/repo-service.ts";
import {Button} from "@/components/ui/button.tsx";
import RepoInfoCard from "@/components/repo-info-card/repo-info-card.tsx";
import Spinner from "@/components/Spinner.tsx";
import {clsx} from "clsx";
import styles from "./repo.module.scss";
import {PackagePlus} from "lucide-react";
import Link from "@/components/Link.tsx";

const RepoList = () => {
    const repoListQuery = useQuery({
        queryKey: ["repo-list"],
        queryFn: listRepos,
    });

    const repos = repoListQuery.data?.data;

    if (repoListQuery.isError) {
        return <Navigate to={"/error"}
                         state={{message: "An error occurred while fetching the repository list."}}/>;
    }

    if (repoListQuery.isPending || repoListQuery.isRefetching) {
        return (
            <Spinner/>
        );
    }

    if (!repoListQuery.isSuccess || !repos) {
        return <Navigate
            to={"/error"}
            state={{message: "An error occurred while fetching the repository list."}}/>;
    }

    return (
        <>
            <div className={"mb-4 flex justify-end"}>
                <Link to={"/repo/setup/postgres"}>
                    <Button size={"sm"} variant={"outline"}>
                        <PackagePlus
                            size={13}
                            style={{width: 15, height: 15, marginRight: "-2px"}}
                            className={"relative top-[-0.5px]"}/>
                        New Repository
                    </Button>
                </Link>
            </div>

            {!!repos && !!repos.length && (
                <div className={"mb-12"}>
                    <div className={clsx("grid gap-6", styles.repoGrid)}>
                        {repos.map(repo => (
                            <RepoInfoCard key={repo.id} repo={repo}/>
                        ))}
                    </div>
                </div>
            )}

            {!!repos && !repos.length && (
                <div className={"mt-40 flex flex-col items-center"}>
                    <img src={"/images/purrcy_confused.png"} width={"350px"} alt={"Purrcy is confused"}/>
                    <p className={"text-center"}>No repositories found. Create one?</p>
                </div>
            )}
        </>
    )
}

export default RepoList
