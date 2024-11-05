import {Link, useLocation} from "react-router-dom";
import {Button} from "@/components/ui/button.tsx";
import {ArrowLeft} from "lucide-react";
import React from "react";

interface ErrorState {
    message: string;
}

const Error = () => {
    const error = useLocation()?.state as ErrorState;

    console.log(error);
    return (
        <div>
            <h1 className={"text-3xl"}>Oops!</h1>
            <p className={"mt-2"}>{error ? error.message : "Something went wrong."}</p>

            <Link to={`/`} className={"mt-8 block"}>
                <Button>
                    <ArrowLeft/> Dashboard
                </Button>
            </Link>
        </div>
    );
};

export default Error;
