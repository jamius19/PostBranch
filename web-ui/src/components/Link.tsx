import {Link as RRLink, LinkProps as RRLinkProps} from "react-router-dom";
import React from "react";


interface LinkProps extends RRLinkProps {
    children: React.ReactNode;
    disabled?: boolean;
}

const Link = (props: LinkProps) => {
    const {children, disabled = false, ...rest} = props;

    return (
        <div>
            <RRLink
                {...rest}
                style={{pointerEvents: disabled ? "none" : "auto"}}>
                {children}
            </RRLink>
        </div>
    );
};

export default Link;
