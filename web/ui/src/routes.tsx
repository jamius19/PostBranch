import {createBrowserRouter} from "react-router-dom";
import RepoList from "@/route/repo/repo-list.tsx";
import Layout from "./layout/layout.tsx";
import Error from "./route/error/error.tsx";
import Repo from "@/route/repo/repo.tsx";
import RepoSetup from "@/route/repo/setup/repo-setup.tsx";
import PgSetupHost from "@/route/repo/setup/pg-adapter/pg-setup-host.tsx";
import PgImportMode from "@/route/repo/setup/pg-import-mode.tsx";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Layout/>,
        errorElement: <Layout><Error/></Layout>,
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
                path: "/repo/:repoName",
                element: <Repo/>,
            },
            {
                path: "/repo/setup/postgres",
                element: <PgImportMode/>,
            },
            {
                path: "/repo/setup/postgres/host",
                element: <PgSetupHost/>,
            },
            {
                path: "/repo/setup/:repoName/postgres",
                element: <PgImportMode/>,
            },
            {
                path: "/repo/setup/postgres/:repoName/host",
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
