import {LoaderCircle} from "lucide-react";

const Spinner = (props: { isLoading?: boolean }) => {
    const {isLoading = true} = props;

    return (
        <>
            {isLoading && <LoaderCircle className={"animate-spin"}/>}
        </>
    );
};

export default Spinner;