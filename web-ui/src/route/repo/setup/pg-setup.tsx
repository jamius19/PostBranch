import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {Navigate, useNavigate, useParams} from "react-router-dom";


const PgSetup = () => {
    const {repoId: repoIdStr} = useParams<{ repoId: string }>();
    const repoId = parseInt(repoIdStr!);


    if (!repoId) {
        return <Navigate to={"/error"} />;
    }


    return (
        <div>
            <div className={"mb-3 cursor-default"}>
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            Configure Storage
                        </BreadcrumbItem>
                        <BreadcrumbSeparator>
                            <ChevronRightIcon/>
                        </BreadcrumbSeparator>
                        <BreadcrumbItem>
                            <BreadcrumbPage>Import Postgres</BreadcrumbPage>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator>
                            <ChevronRightIcon/>
                        </BreadcrumbSeparator>
                        <BreadcrumbItem>
                            Done
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <h1 className={"mb-10"}>Repository Setup</h1>

        </div>
    );
};

export default PgSetup;