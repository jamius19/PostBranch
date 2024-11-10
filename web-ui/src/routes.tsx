import {createBrowserRouter} from "react-router-dom";
import RepoList from "@/route/repo/repo-list.tsx";
import Layout from "./layout/layout.tsx";
import Error from "./route/error/error.tsx";
import PgSetupLocal from "@/route/repo/setup/pg/pg-setup-local.tsx";
import Repo from "@/route/repo/[id]/repo.tsx";
import RepoSetup from "@/route/repo/setup/repo-setup.tsx";
import PgSetupHost from "@/route/repo/setup/pg/pg-setup-host.tsx";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Layout/>,
        errorElement: <Error/>,
        children: [
            {
                path: "/",
                element: <RepoList/>,
            },
            {
                path: "/repo/setup/storage",
                element: <RepoSetup/>,
            },
            {
                path: "/repo/:repoId",
                element: <Repo/>,
            },
            {
                path: "/repo/setup/postgres/local",
                element: <PgSetupLocal/>,
            },
            {
                path: "/repo/setup/postgres/host",
                element: <PgSetupHost/>,
            },
            {
                path: "error",
                element: <Error/>
            }
        ]
    },
]);

export default router;
