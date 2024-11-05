import {createBrowserRouter} from "react-router-dom";
import RepoList from "@/route/repo/repo-list.tsx";
import Layout from "./layout/layout.tsx";
import RepoSetup from "@/route/repo/setup/repo-setup.tsx";
import Error from "./route/error/error.tsx";
import PgSetup from "@/route/repo/setup/pg-setup.tsx";
import Repo from "@/route/repo/[id]/repo.tsx";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Layout/>,
        children: [
            {
                path: "/",
                element: <RepoList/>,
            },
            {
                path: "/repo/setup",
                element: <RepoSetup/>,
            },
            {
                path: "/repo/:repoId",
                element: <Repo/>,
            },
            {
                path: "/repo/setup/:repoId/postgres",
                element: <PgSetup/>,
            },
            {
                path: "error",
                element: <Error/>
            }
        ]
    },
]);

export default router;
