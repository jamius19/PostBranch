import {Link as RRLink, LinkProps as RRLinkProps} from "react-router-dom";
import React from "react";
import {clsx} from "clsx";


interface LinkProps extends RRLinkProps {
    children: React.ReactNode;
    disabled?: boolean;
}

const Link = (props: LinkProps) => {
    const {children, disabled = false, className, ...rest} = props;

    return (
        <RRLink
            {...rest}
            className={clsx(className, disabled ? "cursor-not-allowed" : "cursor-pointer")}
            style={{pointerEvents: disabled ? "none" : "auto"}}>
            {children}
        </RRLink>
    );
};

export default Link;
