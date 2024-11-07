import {Link, useLocation} from "react-router-dom";
import React from "react";

interface ErrorState {
    message?: string;
}

const Error = () => {
    const error = useLocation()?.state as ErrorState;

    return (
        <div className={"mt-40 flex flex-col items-center"}>
            <img src={"/images/purrcy_sending_love.png"} width={"320px"} alt={"Purrcy is confused"}/>
            <h1 className={"text-3xl mt-4"}>Oops!</h1>
            <p className={"mt-0"}>
                {error.message ?? "Something went wrong."}&nbsp;
                <Link to={`/`} className={"text-blue-600 inline-block"}>
                    Go back to Home.
                </Link>
            </p>
        </div>
    );
};

export default Error;
