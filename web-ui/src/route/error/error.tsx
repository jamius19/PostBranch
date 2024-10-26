import {Link} from "react-router-dom";

const Error = () => {
    return (
        <div>
            <h1 className={"text-3xl"}>Oops!</h1>
            <p className={"mt-2"}>Something went wrong. Please go back to <Link to={"/"}>Home</Link>.</p>
        </div>
    );
};

export default Error;