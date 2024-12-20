import {clsx} from "clsx";
import {Link, Outlet} from "react-router-dom";
import {Slide, ToastContainer} from "react-toastify";
import React from "react";

interface LayoutProps {
    children?: React.ReactNode
}

const Layout = ({children}: LayoutProps) => {
    return (
        <div className={clsx("max-w-[930px] 3xl:xl:max-w-[1040px] mx-auto mt-20 mb-20")}>
            <header className={"mb-14"}>
                <Link to="/">
                    <img src={"/postbranch_logo.png"} alt="PostBranch Logo" className={"w-auto h-[45px]"}/>
                </Link>
            </header>
            <ToastContainer
                position="bottom-right"
                autoClose={5000}
                hideProgressBar={false}
                newestOnTop={false}
                closeOnClick
                rtl={false}
                pauseOnFocusLoss
                draggable
                pauseOnHover
                theme="light"
                transition={Slide}
            />
            {children ? children : <Outlet/>}
        </div>
    )
}

export default Layout;
