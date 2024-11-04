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
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {generateName} from '@criblinc/docker-names'
import Spinner from "@/components/Spinner.tsx";
import {ArrowRight, Check} from "lucide-react";
import {getRandomInt} from "@/util/lib.ts";
import React, {useCallback, useEffect, useMemo, useState} from "react";
import {Link} from "react-router-dom";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import StorageSlider from "@/components/storage-slider.tsx";
import {useAppForm} from "@/lib/hooks/use-app-form.ts";

const blockSchema = z.object({
    repoType: z.literal("block"),
});

const virtualSchema = z.object({
    repoType: z.literal("virtual"),
    sizeInMb: z.number({message: "Size value must be in the format (550MB, 3.5GB, 1TB)"})
        .min(256, "Minimum size value is 256 Megabytes")
});

const RepoSetup = () => {
    // TODO: Show block storage list somewhere
    // const {isSuccess, data} = useQuery({
    //     queryKey: ["repo-block-storage"],
    //     queryFn: listBlockStorages
    // });

    const reposQuery = useQuery({
        queryKey: ["repo-list"],
        queryFn: listRepos,
    });

    const repoInit = useNotifiableMutation({
        mutationKey: ["repo-init"],
        mutationFn: initRepo,
        messages: {
            pending: "Creating repository",
            success: "Repository created successfully",
        },
        invalidates: ["repo-list"],
    });

    const [nameUpdated, setNameUpdated] = useState(false)

    const generatedName = generateName();
    const repoInitSuccess = repoInit.isSuccess;
    const repoInitPending = repoInit.isPending;

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
            .refine(val => !reposQuery.data?.data?.some(repo => repo.path === val),
                "Repository with the same path already exists"),
    }), [reposQuery]);

    const formSchema = useMemo(
        () => z.discriminatedUnion("repoType", [virtualSchema, blockSchema]).and(baseFormSchema),
        [baseFormSchema]
    );

    const defaultFormValues = useMemo<RepoInitDto>(() => ({
        name: generatedName,
        repoType: "virtual",
        path: getVirtualPath(generatedName),
        sizeInMb: 1024,
    }), [generatedName]);

    const repoForm = useAppForm<RepoInitDto>({
        resolver: zodResolver(formSchema),
        defaultValues: defaultFormValues,
        mode: "onChange"
    });

    const repoType = repoForm.watch("repoType");

    const onSubmit = useCallback(async (data: RepoInitDto) => {
        console.log(data);
        await repoInit.mutateAsync(data);
    }, [repoInit]);

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

    if (reposQuery.isPending || !nameUpdated) {
        return <Spinner/>;
    }

    return (
        <div>
            <div className={"mb-3 cursor-default"}>
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            <BreadcrumbPage>Configure Storage</BreadcrumbPage>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator>
                            <ChevronRightIcon/>
                        </BreadcrumbSeparator>
                        <BreadcrumbItem>
                            Import Data
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <h1 className={"mb-10"}>Repository Setup</h1>

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

                    {repoType === 'virtual' && (
                        <div className="flex gap-4">
                            <StorageSlider
                                disabled={repoInitPending || repoInitSuccess}
                                name={"sizeInMb"}/>
                        </div>
                    )}

                    <div className={"flex gap-4"}>
                        <Button
                            type="submit"
                            variant={repoInitSuccess ? "success" : "default"}
                            disabled={repoInitPending || repoInitSuccess}>

                            <Spinner isLoading={repoInitPending}/>
                            {repoInitSuccess && <Check/>}
                            {repoInitSuccess ? "Repository created" : "Create repository"}
                        </Button>

                        {repoInitSuccess && (
                            <Link to={`/repo/setup/${repoInit.data.data!.id}/postgres`}>
                                <Button>
                                    Setup Postgres <ArrowRight/>
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