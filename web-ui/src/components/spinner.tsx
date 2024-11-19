import {LoaderCircle} from "lucide-react";

const Spinner = (
    props: { isLoading?: boolean, size?: number, light?: boolean }
) => {
    
    const {isLoading = true, size = 24, light = false} = props;

    return (
        <>
            {isLoading &&
                <LoaderCircle
                    size={size}
                    className={"animate-spin"}
                    color={light ? "#ffffff" : "#1F2937"}/>}
        </>
    );
};

export default Spinner;
