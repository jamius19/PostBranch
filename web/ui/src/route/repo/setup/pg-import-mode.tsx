import React, {useState} from "react";
import {ArrowRight, Network} from "lucide-react";
import {PgAdapterName} from "@/@types/repo/pg/pg-response-dto.ts";
import {twMerge as tm} from "tailwind-merge";
import Link from "@/components/link.tsx";
import {Button} from "@/components/ui/button.tsx";
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {useParams} from "react-router-dom";

const SelectedBgClass = "border border-gray-700 bg-gray-700 text-white shadow-gray-200 shadow-lg";
const NormalBgClass = "border border-gray-300 hover:bg-gray-100/80";


const PgImportMode = () => {
    const {repoName} = useParams<{ repoName: string }>();

    const [selectedType, setSelectedType] = useState<PgAdapterName | undefined>();

    const getNextLink = () => {
        if (!repoName) {
            return `/repo/setup/postgres/${selectedType}`;
        } else {
            return `/repo/setup/postgres/${repoName}/${selectedType}`;
        }
    }

    return (
        <div>
            <div className={"mb-3 cursor-default"}>
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            <BreadcrumbPage>Configure Connection</BreadcrumbPage>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator>
                            <ChevronRightIcon/>
                        </BreadcrumbSeparator>
                        <BreadcrumbItem>
                            Configure Postgres
                        </BreadcrumbItem>
                        <BreadcrumbSeparator>
                            <ChevronRightIcon/>
                        </BreadcrumbSeparator>
                        <BreadcrumbItem>
                            Configure Storage
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>

            <h1 className={"mb-10"}>Connection Type</h1>

            <p className={"mb-8 text-sm"}>
                Choose what type of connection will be used to connect with your Postgres instance.<br/>
                PostBranch will import the data from there into the repository.
            </p>

            <div className={"flex flex-col gap-6 w-[550px] select-none"}>
                <div
                    onClick={() => setSelectedType("host")}
                    className={tm("cursor-pointer px-8 py-6 rounded-lg shadow-gray-100 hover:shadow-gray-200 shadow-md hover:shadow-lg transition-all duration-200 flex flex-row gap-4",
                        selectedType !== "host" && NormalBgClass,
                        selectedType === "host" && SelectedBgClass)}>
                    <Network className={"flex-shrink-0 relative top-[3px]"} size={20}/>
                    <div>
                        <div className={"flex gap-2"}>
                            <p className={"font-bold"}>Host Connection</p>
                        </div>
                        <p className={"text-sm mt-2.5"}>
                            Choose this if your Postgres server is running on a remote
                            machine or you want to connect via <code className={"font-bold"}>host</code> based
                            authentication with username, and password
                        </p>
                    </div>
                </div>
            </div>

            <Link disabled={!selectedType} to={getNextLink()} className={"mt-10 block"}>
                <Button disabled={!selectedType}>
                    Configure Postgres <ArrowRight/>
                </Button>
            </Link>
        </div>
    );
};

export default PgImportMode;
