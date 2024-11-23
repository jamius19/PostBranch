import {z} from "zod";
import {zodResolver} from "@hookform/resolvers/zod";
import {Input} from "@/components/ui/input.tsx";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select.tsx";
import {Button} from "@/components/ui/button.tsx";
import {
    Form,
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage
} from "@/components/ui/form.tsx";
import {RepoInitDto, RepoType} from "@/@types/repo/repo-init-dto.ts";
import {useQuery} from "@tanstack/react-query";
import {initRepo, listRepos} from "@/service/repo-service.ts";
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {generateName} from '@criblinc/docker-names'
import Spinner from "@/components/spinner.tsx";
import {ArrowRight, Check, Info} from "lucide-react";
import {formatValue, getRandomInt} from "@/util/lib.ts";
import React, {JSX, useCallback, useEffect, useMemo, useState} from "react";
import {Link, Navigate, useLocation} from "react-router-dom";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import StorageSlider, {MIN_VALUE} from "@/components/storage-slider.tsx";
import {useAppForm} from "@/lib/hooks/use-app-form.ts";
import {PgAdapterName, PgResponseDto} from "@/@types/repo/pg/pg-response-dto.ts";
import {ChevronRightIcon} from "@radix-ui/react-icons";

const blockSchema = z.object({
    repoType: z.literal("block"),
});

interface RepoSetupState {
    pgResponse?: PgResponseDto;
    adapter?: PgAdapterName
}

const RepoSetup = (): JSX.Element => {
    const repoSetupState = useLocation()?.state as RepoSetupState;

    const reposQuery = useQuery({
        queryKey: ["repo-list"],
        queryFn: listRepos,
    });

    const repoInit = useNotifiableMutation({
        mutationKey: ["repo-init"],
        mutationFn: (repoConfig: RepoInitDto) => {
            return initRepo(
                {repoConfig, pgConfig: repoSetupState.pgResponse!.pgConfig},
                repoSetupState.adapter!
            );
        },
        messages: {
            pending: "Creating repository",
            success: "Repository created and Postgres import started",
        },
        invalidates: ["repo-list"],
    });

    const [nameUpdated, setNameUpdated] = useState(false)

    const generatedName = generateName();
    const repoInitSuccess = repoInit.isSuccess;
    const repoInitPending = repoInit.isPending;
    const repoSizeMinValue = repoSetupState.pgResponse!.clusterSizeInMb + MIN_VALUE;

    const baseFormSchema = useMemo(() => z.object({
        name: z.string()
            .min(1, "Name is required")
            .max(50, "Name must be less than 50 characters")
            .regex(/^[a-z][a-z0-9-]*[a-z0-9]$/, "Name must start with a letter, end with a letter or number, and can only contain letters, numbers, and hyphens")
            .refine(val => !reposQuery.data?.data?.some(repo => repo.name === val),
                "Repository with the same name already exists"),
        path: z.string()
            .min(1, "Path is required")
            .max(2000, "Path must be less than 2000 characters")
            .refine(value => value.startsWith("/") && !value.endsWith("/"), {
                message: "Path must start with '/' and not end with '/'",
            })
            .refine(value => !value.includes(" "), {
                message: "Path path must not contain spaces",
            })
            .refine(val => !reposQuery.data?.data?.some(repo => repo.pool.path === val),
                "Repository with the same path already exists"),
    }), [reposQuery]);

    const virtualSchema = useMemo(() => z.object({
        repoType: z.literal("virtual"),
        sizeInMb: z.number({message: "Repository Size must be in the format (550MB, 3.5GB, 1TB)"})
            .min(repoSizeMinValue, `Minimum allowed size is ${formatValue(repoSizeMinValue)}`)
    }), [repoSizeMinValue]);

    const formSchema = useMemo(
        () => z.discriminatedUnion("repoType", [virtualSchema, blockSchema]).and(baseFormSchema),
        [baseFormSchema, virtualSchema]
    );

    const defaultFormValues = useMemo<RepoInitDto>(() => ({
        name: generatedName,
        repoType: "virtual",
        path: getVirtualPath(generatedName),
        sizeInMb: repoSizeMinValue,
    }), [generatedName, repoSizeMinValue]);

    const repoForm = useAppForm<RepoInitDto>({
        resolver: zodResolver(formSchema),
        defaultValues: defaultFormValues,
        mode: "onChange"
    });

    const repoType = repoForm.watch("repoType");

    const onSubmit = useCallback(async (data: RepoInitDto) => {
        await repoInit.mutateAsync(data);
    }, [repoInit, repoSetupState?.pgResponse]);

    const clearVirtualStorageValues = useCallback((value: RepoType) => {
        if (value === 'block') {
            repoForm.setValue("path", "")
        } else {
            repoForm.setValue("path", getVirtualPath(repoForm.getValues().name));
        }
    }, [repoForm]);

    useEffect(() => {
        if (reposQuery.isSuccess) {
            const needsNewName = reposQuery
                .data
                .data
                ?.some(repo => repo.name === generatedName);

            if (needsNewName) {
                let newName = generateName();

                while (reposQuery.data.data?.some(repo => repo.name === newName)) {
                    newName = generateName();
                }

                repoForm.setValue("name", newName);
                repoForm.setValue("path", getVirtualPath(newName));
            }

            setNameUpdated(true);
        }
    }, [generatedName, repoForm, reposQuery.data, reposQuery.isSuccess]);

    if (!repoSetupState.pgResponse || !repoSetupState.adapter) {
        return <Navigate to={"/error"} state={{message: "No Postgres configuration found."}}/>;
    }

    if (reposQuery.isPending || !nameUpdated) {
        return <Spinner/>;
    }

    return (
        <div>
            <div className={"mb-3 cursor-default"}>
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            Configure Connection
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
                            <BreadcrumbPage>Configure Storage</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <h1 className={"mb-10"}>Repository Storage Setup</h1>

            <Form {...repoForm}>
                <form
                    onSubmit={repoForm.handleSubmit(onSubmit)}
                    onKeyDown={repoForm.disableSubmit}
                    className="w-2/3 space-y-8">

                    <FormField
                        control={repoForm.control}
                        name="name"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Repository Name</FormLabel>
                                <FormControl>
                                    <Input {...field}
                                           disabled={repoInitPending || repoInitSuccess}
                                           placeholder="Enter Repository Name"
                                           spellCheck="false"
                                           onChange={e => {
                                               const val: string = e.target.value.toLowerCase().trim();
                                               field.onChange(val);
                                           }}/>
                                </FormControl>
                                <FormDescription>
                                    Enter the name of this repository
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={repoForm.control}
                        name="repoType"
                        render={({field}) => (
                            <FormItem className="flex-1">
                                <FormLabel>Storage Type</FormLabel>
                                <Select
                                    disabled={repoInitPending || repoInitSuccess}
                                    defaultValue={field.value}
                                    onValueChange={(value: RepoType) => {
                                        field.onChange(value);
                                        clearVirtualStorageValues(value);
                                    }}>
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select repository type"/>
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="block">Block Storage</SelectItem>
                                        <SelectItem value="virtual">Virtual Storage</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormDescription>
                                    <span>Select the desired storage type for the repository</span><br/>
                                    <span className={"text-zinc-400"}>
                                        If you don&apos;t have any external drive, choose the <b>Virtual
                                        Storage</b> option</span>
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={repoForm.control}
                        name="path"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>
                                    {repoType === 'virtual' ? "Repository Path" : "Block Storage Path"}
                                </FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        disabled={repoInitPending || repoInitSuccess}
                                        className={"mono"}
                                        placeholder={repoType === 'virtual' ? "Enter repository path" : "/dev/vdb"}
                                        spellCheck="false"
                                        onChange={e => {
                                            const val: string = e.target.value.trim();
                                            field.onChange(val);
                                        }}/>
                                </FormControl>
                                <FormDescription>
                                    Enter the path&nbsp;
                                    {repoType === 'virtual' ? 'for the virtual repository' : 'of the block storage'}
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <div
                        className={"flex items-center gap-3 bg-blue-200/70 text-xs text-blue-700 rounded-md px-4 py-3"}>
                        <Info size={16} className={"relative bottom-[1px] flex-shrink-0"}/>

                        <p>
                            <b>Minimum required storage space to clone the database cluster
                                is {formatValue(repoSizeMinValue)}</b>
                        </p>
                    </div>

                    {repoType === "virtual" && (
                        <div className="flex flex-col">
                            <StorageSlider
                                formProps={{
                                    disabled: repoInitPending || repoInitSuccess,
                                    name: "sizeInMb",
                                }}
                                minSizeInMb={repoSetupState.pgResponse.clusterSizeInMb}/>
                        </div>
                    )}

                    <div className={"flex gap-4"}>
                        <Button
                            type="submit"
                            variant={repoInitSuccess ? "success" : "default"}
                            disabled={repoInitPending || repoInitSuccess}>

                            <Spinner isLoading={repoInitPending} light/>
                            {repoInitSuccess && <Check/>}
                            {repoInitSuccess ? "Repository created" : "Create repository"}
                        </Button>

                        {repoInitSuccess && (
                            <Link to={`/repo/${repoInit.data.data!.name}`}>
                                <Button>
                                    Go to Repository <ArrowRight/>
                                </Button>
                            </Link>
                        )}
                    </div>
                </form>
            </Form>

        </div>
    );
};

const getVirtualPath = (generatedName: string) => {
    return `/var/lib/post-branch/${generatedName}-${getRandomInt(0, 100)}.img`;
}

export default RepoSetup;
