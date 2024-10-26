import {createBrowserRouter} from "react-router-dom";
import Repo from "@/route/repo/repo.tsx";
import Layout from "./layout/layout.tsx";
import RepoSetup from "@/route/repo/setup/repo-setup.tsx";
import Error from "./route/error/error.tsx";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Layout/>,
        children: [
            {
                path: "/",
                element: <Repo/>,
            },
            {
                path: "/repo/setup",
                element: <RepoSetup/>,
            },
            {
                path: "error",
                element: <Error/>
            }
        ]
    },
]);

export default router;