import {StrictMode} from 'react';
import {createRoot} from 'react-dom/client';
import {RouterProvider,} from "react-router-dom";
import {QueryClient, QueryClientProvider,} from '@tanstack/react-query'
import 'react-toastify/dist/ReactToastify.css';
import './main.css';
import router from "@/routes.tsx";
import {ReactQueryDevtools} from "@tanstack/react-query-devtools";

const queryClient = new QueryClient()

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <RouterProvider router={router}/>
            <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
    </StrictMode>,
)
