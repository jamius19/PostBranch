import {LoaderCircle} from "lucide-react";

const Spinner = (props: { isLoading?: boolean, size?: number }) => {
    const {isLoading = true, size = 24} = props;

    return (
        <>
            {isLoading && <LoaderCircle size={size} className={"animate-spin"} color={"#63A8F5"}/>}
        </>
    );
};

export default Spinner;
