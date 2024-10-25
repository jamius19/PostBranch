import React from "react";
import {clsx} from "clsx";
import styles from './layout.module.scss';

const Layout = ({children}: { children: React.ReactNode }) => {
    return (
        <div className={clsx(styles.container, "max-w-[930px] 2xl:max-w-[1140px] mx-auto mt-20")}>
            <header className={"mb-8"}>
                <img src={"/postbranch_logo.png"} alt="PostBranch Logo" className={"w-auto h-[40px]"}/>
            </header>
            {children}
        </div>
    )
}

export default Layout;