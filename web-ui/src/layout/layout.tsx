import {clsx} from "clsx";
import {Link, Outlet} from "react-router-dom";
import {Slide, ToastContainer} from "react-toastify";

const Layout = () => {
    return (
        <div className={clsx("max-w-[930px] mx-auto mt-20 mb-20")}>
            <header className={"mb-14"}>
                <Link to="/">
                    <img src={"/postbranch_logo.png"} alt="PostBranch Logo" className={"w-auto h-[50px]"}/>
                </Link>
            </header>
            <ToastContainer
                position="bottom-left"
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
            <Outlet/>
        </div>
    )
}

export default Layout;
