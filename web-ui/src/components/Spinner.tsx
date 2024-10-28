import {LoaderCircle} from "lucide-react";

const Spinner = (props: { isLoading?: boolean }) => {
    const {isLoading = true} = props;

    return (
        <>
            {isLoading && <LoaderCircle className={"animate-spin"} color={"#63A8F5"}/>}
        </>
    );
};

export default Spinner;