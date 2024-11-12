import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {z} from "zod";
import {
    Form,
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage
} from "@/components/ui/form.tsx";
import {Input} from "@/components/ui/input.tsx";
import {Button} from "@/components/ui/button.tsx";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select.tsx";
import {zodResolver} from "@hookform/resolvers/zod";
import {useAppForm} from "@/lib/hooks/use-app-form.ts";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import {validatePg} from "@/service/repo-service.ts";
import React, {JSX, useCallback} from "react";
import {Checkbox} from "@/components/ui/checkbox.tsx";
import Spinner from "@/components/Spinner.tsx";
import {ArrowRight, Check, Info} from "lucide-react";
import Link from "@/components/Link.tsx";
import {formatValue} from "@/util/lib.ts";
import {PgLocalInitDto} from "@/@types/repo/pg/pg-local-init-dto.ts";

const formSchema = z.object({
    version: z.number()
        .min(15, "Minimum supported PostgreSQL version is 15")
        .max(17, "Maximum supported PostgreSQL version is 17"),
    postgresPath: z.string()
        .min(1, "PostgreSQL path is required")
        .refine(value => !value.includes(" "), {
            message: "PostgreSQL path must not contain spaces",
        })
        .refine(value => value.startsWith("/") && !value.endsWith("/"), {
            message: "PostgreSQL Path must start with '/' and not end with '/'",
        }),
    stopPostgres: z.boolean({message: "Required"}),
    postgresOsUser: z.string()
        .min(1, "PostgreSQL OS user is required")
        .regex(/^[a-z_]([a-z0-9_-]{0,31}|[a-z0-9_-]{0,30}\$)$/, {
            message: "PostgreSQL OS user must be a valid Unix username",
        })
        .refine(value => !value.includes(" "), {
            message: "PostgreSQL OS user must not contain spaces",
        }),
});

const defaultValues: PgLocalInitDto = {
    version: 16,
    postgresPath: "",
    stopPostgres: true,
    postgresOsUser: "postgres",
};

const PgSetupLocal = (): JSX.Element => {
    // const {repoId: repoIdStr} = useParams<{ repoId: string }>();
    // const repoId = parseInt(repoIdStr!);

    const pgValidate = useNotifiableMutation({
        mutationKey: ["pg-import-local"],
        mutationFn: (pgInit: PgLocalInitDto) => validatePg(pgInit, "local"),
        messages: {
            pending: "Checking PostgreSQL configuration",
            success: "PostgreSQL configuration is valid",
        },
        invalidates: ["repo-list", "repo"],
    });

    const pgForm = useAppForm({
        defaultValues: defaultValues,
        resolver: zodResolver(formSchema),
    });

    const onSubmit = useCallback(async (pgInit: PgLocalInitDto) => {
        await pgValidate.mutateAsync(pgInit);
    }, [pgValidate]);

    // if (!isInteger(repoIdStr)) {
    //     return <Navigate to={"/error"} state={{message: "The repository ID in the URL is invalid."}}/>;
    // }

    const repoInitSuccess = pgValidate.isSuccess;
    const repoInitPending = pgValidate.isPending;

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
                            <BreadcrumbPage>Configure PostgreSQL</BreadcrumbPage>
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
            <h1 className={"mb-10"}>PostgreSQL Connection</h1>

            <Form {...pgForm}>
                <form
                    onSubmit={pgForm.handleSubmit(onSubmit)}
                    onKeyDown={pgForm.disableSubmit}
                    className="w-2/3 space-y-8">

                    <FormField
                        control={pgForm.control}
                        name="version"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Postgres Version</FormLabel>
                                <Select
                                    disabled={repoInitPending || repoInitSuccess}
                                    defaultValue={String(field.value)}
                                    onValueChange={(value: string) => {
                                        field.onChange(Number(value));
                                    }}>
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select Postgres Version"/>
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="15">15</SelectItem>
                                        <SelectItem value="16">16</SelectItem>
                                        <SelectItem value="17">17</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormDescription>
                                    Select the version of PostgreSQL
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={pgForm.control}
                        name="postgresPath"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Postgres Installation Path</FormLabel>
                                <FormControl>
                                    <Input {...field}
                                           disabled={repoInitPending || repoInitSuccess}
                                           spellCheck="false"
                                           placeholder={"/usr/lib/postgresql/16"}
                                           onChange={e => {
                                               const val: string = e.target.value.trim();
                                               field.onChange(val);
                                           }}/>
                                </FormControl>
                                <FormDescription>
                                    Absolute path to PostgreSQL installation
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={pgForm.control}
                        name="stopPostgres"
                        render={({field}) => (
                            <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                                <FormControl>
                                    <Checkbox
                                        disabled={repoInitPending || repoInitSuccess}
                                        checked={field.value}
                                        onCheckedChange={field.onChange}
                                    />
                                </FormControl>
                                <div className="space-y-1 leading-none">
                                    <FormLabel>
                                        Stop Current PostgreSQL after importing data (Recommended)
                                    </FormLabel>
                                    <FormDescription className={"leading-5"} style={{marginTop: "0.4rem"}}>
                                        PostBranch will automatically start Postgres with identical data after importing
                                        it
                                    </FormDescription>
                                </div>
                            </FormItem>
                        )}
                    />

                    <hr className={"my-10 border-dotted"}/>

                    <FormField
                        control={pgForm.control}
                        name="postgresOsUser"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Postgres OS User</FormLabel>
                                <FormControl>
                                    <Input {...field}
                                           disabled={repoInitPending || repoInitSuccess}
                                           spellCheck="false"
                                           placeholder={"postgres"}
                                           onChange={e => {
                                               const val: string = e.target.value.trim();
                                               field.onChange(val);
                                           }}/>
                                </FormControl>
                                <FormDescription>
                                    Username of the Local PostgreSQL operating system account.<br/>
                                    If you&#39;re unsure, use the default value <code
                                    className={"font-bold"}>postgres</code><br/>
                                    <i>
                                        (This MUST be a <code
                                        className={"font-bold"}>superuser</code>)
                                    </i>
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    {repoInitSuccess && (
                        <div
                            className={"flex items-center gap-2 bg-lime-200/70 text-xs text-lime-700 rounded-md px-4 py-3"}>
                            <Info size={16} className={"relative bottom-[1px] flex-shrink-0"}/>

                            <p>
                                <b>Connection successful! The database cluster size
                                    is {formatValue(pgValidate.data.data!.clusterSizeInMb)}</b>.
                            </p>
                        </div>
                    )}

                    <div className={"flex gap-4"}>
                        <Button
                            type="submit"
                            variant={repoInitSuccess ? "success" : "default"}
                            disabled={repoInitPending || repoInitSuccess}>

                            <Spinner isLoading={repoInitPending}/>
                            {repoInitSuccess && <Check/>}
                            {repoInitSuccess ? "Connected" : "Connect"}
                        </Button>

                        {repoInitSuccess && (
                            <Link to={"/repo/setup/storage"}
                                  state={{pgConfig: pgValidate.data.data!, adapter: "local"}}>
                                <Button>
                                    Storage Configuration <ArrowRight/>
                                </Button>
                            </Link>
                        )}
                    </div>
                </form>
            </Form>
        </div>
    );
};

export default PgSetupLocal;
