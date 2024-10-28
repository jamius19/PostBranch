import {z} from "zod";
import {useForm} from "react-hook-form";
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
import {useMutation, useQuery} from "@tanstack/react-query";
import {initRepo, listRepoNames, listRepos} from "@/service/repo-service.ts";
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList, BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {generateName} from '@criblinc/docker-names'
import Spinner from "@/components/Spinner.tsx";
import {ArrowRight, Check} from "lucide-react";
import {getRandomInt} from "@/util/lib.ts";
import {useEffect, useRef, useState} from "react";
import {Link} from "react-router-dom";
import {Slide, toast} from "react-toastify";

const virtualSchema = z.object({
    repoType: z.literal("virtual"),
    size: z.number().positive("Size value must be greater than 0"),
    sizeUnit: z.enum(["K", "M", "G"])
});

const blockSchema = z.object({
    repoType: z.literal("block"),
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

    const repoUpdate = useMutation({
        mutationFn: initRepo
    });

    const [nameUpdated, setNameUpdated] = useState(false)

    const generatedName = generateName();
    const repoInitSuccess = repoUpdate.isSuccess;
    const repoInitPending = repoUpdate.isPending;

    const baseFormSchema = z.object({
        name: z.string()
            .min(1, "Name is required")
            .regex(/^[a-z][a-z0-9-]*[a-z0-9]$/, "Name must start with a letter, end with a letter or number, and can only contain letters, numbers, and hyphens")
            .refine(val => !reposQuery.data?.data?.some(repo => repo.name === val),
                "Repository with the same name already exists"),
        path: z.string()
            .min(1, "Path is required")
            .regex(/^\/\S+$/, "Path must start with / and cannot contain spaces")
            .refine(val => !reposQuery.data?.data?.some(repo => repo.path === val),
                "Repository with the same path already exists"),
    });

    const formSchema = z.discriminatedUnion("repoType", [virtualSchema, blockSchema])
        .and(baseFormSchema);

    const defaultFormValues: RepoInitDto = {
        name: generatedName,
        repoType: "virtual",
        path: getVirtualPath(generatedName),
        size: 100,
        sizeUnit: "M"
    };

    const form = useForm<RepoInitDto>({
        resolver: zodResolver(formSchema),
        defaultValues: defaultFormValues,
        mode: "onChange"
    });

    const repoType = form.watch("repoType");

    const onSubmit = async (data: RepoInitDto) => {
        console.log(data);
        repoUpdate.mutate(data);
    };

    const clearVirtualStorageValues = (value: RepoType) => {
        if (value === 'block') {
            form.setValue("size", undefined);
            form.setValue("sizeUnit", undefined);
            form.setValue("path", "")
        } else {
            form.setValue("size", 1);
            form.setValue("sizeUnit", 'G');
            form.setValue("path", "/var/lib/post-branch/virtualdisk01.img")
        }
    }

    useEffect(() => {
        if (reposQuery.isSuccess) {
            const needsNewName = reposQuery.data.data?.some(repo => repo.name === generatedName);

            if (needsNewName) {
                const newName = generateName();
                form.setValue("name", newName);
                form.setValue("path", getVirtualPath(newName));
            }

            setNameUpdated(true);
        }
    }, [reposQuery.isSuccess]);

    useEffect(() => {
        if (repoUpdate.isError) {
            toast.error(repoUpdate.error.message);
        }

    }, [repoUpdate.isError]);

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
                            Import Postgres
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

            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="w-2/3 space-y-8">

                    <FormField
                        disabled={repoInitPending || repoInitSuccess}
                        control={form.control}
                        name="name"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Repository Name</FormLabel>
                                <FormControl>
                                    <Input {...field}
                                           placeholder="Enter Repository Name"
                                           spellCheck="false"
                                           onChange={e => {
                                               e.target.value = e.target.value.toLowerCase().trim();
                                               field.onChange(e);
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
                        control={form.control}
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
                                        If you don't have any external drive, choose the <b>Virtual
                                        Storage</b> option</span>
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="path"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>{repoType === 'virtual' ? "Repository Path" : "Block Storage Path"}</FormLabel>
                                <FormControl>
                                    <Input
                                        {...field}
                                        disabled={repoInitPending || repoInitSuccess}
                                        className={"mono"}
                                        placeholder={repoType === 'virtual' ? "Enter repository path" : "/dev/vdb"}
                                        spellCheck="false"
                                        onChange={e => {
                                            e.target.value = e.target.value.trim();
                                            field.onChange(e);
                                        }}/>
                                </FormControl>
                                <FormDescription>
                                    Enter the
                                    path {repoType === 'virtual' ? 'for the virtual repository' : 'of the block storage'}
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    {repoType === 'virtual' && (
                        <div className="flex gap-4">
                            <FormField
                                disabled={repoInitPending || repoInitSuccess}
                                control={form.control}
                                name="size"
                                render={({field}) => (
                                    <FormItem className="flex-1">
                                        <FormLabel>Size</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                placeholder="Enter size"
                                                {...field}
                                                onChange={e => field.onChange(Number(e.target.value))}
                                            />
                                        </FormControl>
                                        <FormDescription>
                                            Enter the size of the virtual disk
                                        </FormDescription>
                                        <FormMessage/>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="sizeUnit"
                                render={({field}) => (
                                    <FormItem className="flex-1">
                                        <FormLabel>Size Unit</FormLabel>
                                        <Select
                                            disabled={repoInitPending || repoInitSuccess}
                                            onValueChange={field.onChange}
                                            defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select unit"/>
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="K">KB</SelectItem>
                                                <SelectItem value="M">MB</SelectItem>
                                                <SelectItem value="G">GB</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormDescription>
                                            Select the unit for the disk size
                                        </FormDescription>
                                        <FormMessage/>
                                    </FormItem>
                                )}
                            />
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
                            <Link to={`/repo/setup/${repoUpdate.data.data!.id}/postgres`}>
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