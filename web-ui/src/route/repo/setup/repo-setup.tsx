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
import {RepoInitDto, RepoType} from "@/@types/repo-init-dto.ts";
import {useQuery} from "@tanstack/react-query";
import {listBlockStorages} from "@/service/repo-service.ts";
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList, BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {generateName} from '@criblinc/docker-names'


const baseFormSchema = z.object({
    name: z.string()
        .min(1, "Name is required")
        .regex(/^[a-zA-Z]+(-[a-zA-Z]+)*$/, "Name can only contain alphabets with hyphens in between"),
    branchName: z.string()
        .min(1, "Branch Name is required")
        .regex(/^[a-z][a-z0-9-]*[a-z0-9]$/, "Branch Name must start with a letter, end with a letter or number, and can only contain letters, numbers, and hyphens"),
    path: z.string().min(1, "Path is required"),
});

const virtualSchema = z.object({
    repoType: z.literal("virtual"),
    size: z.number().min(1, "Size value must be greater than or equal to 1"),
    sizeUnit: z.enum(["K", "M", "G"])
});

const blockSchema = z.object({
    repoType: z.literal("block"),
});

const formSchema = z.discriminatedUnion("repoType", [virtualSchema, blockSchema])
    .and(baseFormSchema);

const RepoSetup = () => {
    // TODO: Show block storage list somewhere
    const {isSuccess, data} = useQuery({
        queryKey: ["repo-block-storage"],
        queryFn: listBlockStorages
    });

    const defaultFormValues: RepoInitDto = {
        name: generateName(),
        branchName: "main",
        repoType: "virtual",
        path: "/var/lib/post-branch/virtualdisk01.img",
        size: 1,
        sizeUnit: "G"
    }

    const form = useForm<RepoInitDto>({
        resolver: zodResolver(formSchema),
        defaultValues: defaultFormValues,
        mode: "onChange"
    });

    const repoType = form.watch("repoType");

    const onSubmit = async (data: RepoInitDto) => {
        try {
            // Handle form submission here
            console.log(data);
        } catch (error) {
            console.error('Error submitting form:', error);
        }
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
            <h1>Repository Setup</h1>

            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="mt-10 w-2/3 space-y-8">

                    <FormField
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
                        name="branchName"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Branch Name</FormLabel>
                                <FormControl>
                                    <Input {...field}
                                           placeholder="Enter Branch Name"
                                           spellCheck="false"
                                           onChange={e => {
                                               e.target.value = e.target.value.toLowerCase().trim();
                                               field.onChange(e);
                                           }}/>
                                </FormControl>
                                <FormDescription>
                                    Enter the name of main Postgres branch
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
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
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

                    <div><Button type="submit">Initialize Repository</Button></div>
                </form>
            </Form>

        </div>
    );
};

export default RepoSetup;