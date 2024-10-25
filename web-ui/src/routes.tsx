import {createBrowserRouter} from "react-router-dom";
import Index from "./route/index";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Index/>,
    },
]);

export default router;