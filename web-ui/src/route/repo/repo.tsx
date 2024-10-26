import {useQuery} from "@tanstack/react-query";
import {Link, Navigate, useNavigate} from "react-router-dom";
import {BeatLoader, CircleLoader, PropagateLoader} from "react-spinners";
import {listRepos} from "@/service/repo-service.ts";
import {Button} from "@/components/ui/button.tsx";

const Repo = () => {
    const {isPending, data: response, error} = useQuery({
        queryKey: ["repo-list"],
        queryFn: () => listRepos(),
    });

    if (error) {
        return <Navigate to={"/error"}/>;
    }

    if (isPending) {
        return (
            <BeatLoader color={"#3687d7"}/>
        );
    }

    return (
        <>
            {response.data ? (
                <>
                    Repositories found!
                    {response?.data.map(repo => repo.name)}
                </>
            ) : (
                <>
                    <p>No Postgres Repositories found.</p>
                    <Link to={"/repo/setup"}>
                        <Button className={"mt-2"}>Add</Button>
                    </Link>
                </>
            )}
        </>
    )
}

export default Repo
