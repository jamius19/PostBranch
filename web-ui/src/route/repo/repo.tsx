import {useQuery} from "@tanstack/react-query";
import {Link, Navigate} from "react-router-dom";
import {listRepos} from "@/service/repo-service.ts";
import {Button} from "@/components/ui/button.tsx";
import RepoInfoCard from "@/components/repo-info-card/repo-info-card.tsx";
import Spinner from "@/components/Spinner.tsx";
import {clsx} from "clsx";
import styles from "./repo.module.scss";
import {ArrowRight} from "lucide-react";

const Repo = () => {
    const {isPending, data: response, error} = useQuery({
        queryKey: ["repo-list"],
        queryFn: listRepos,
    });

    if (error) {
        return <Navigate to={"/error"}/>;
    }

    if (isPending) {
        return (
            <Spinner/>
        );
    }

    return (
        <>
            {!!response.data && !!response.data.length && (
                <div className={"mb-12"}>
                    <div className={clsx("grid gap-6", styles.repoGrid)}>
                        {response?.data.map(repo => (
                            <RepoInfoCard key={repo.id} repo={repo}/>
                        ))}
                    </div>
                </div>
            )}


            <div>
                {!!response.data && !response.data.length && (
                    <p className={"mb-4"}>No repositories found.</p>
                )}

                <Link to={"/repo/setup"}>
                    <Button>
                        Create Repository
                        <ArrowRight/>
                    </Button>
                </Link>
            </div>
        </>
    )
}

export default Repo
