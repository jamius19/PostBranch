import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {Link, Navigate, useParams} from "react-router-dom";
import {z} from "zod";
import {RepoPgInitDto} from "@/@types/repo/repo-pg-init-dto.ts";
import {Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage} from "@/components/ui/form";
import {Input} from "@/components/ui/input";
import {Button} from "@/components/ui/button.tsx";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select.tsx";
import {zodResolver} from "@hookform/resolvers/zod";
import {useAppForm} from "@/lib/hooks/use-app-form.ts";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import {importPg} from "@/service/repo-service.ts";
import React, {useCallback} from "react";
import {Checkbox} from "@/components/ui/checkbox.tsx";
import Spinner from "@/components/Spinner.tsx";
import {ArrowRight, Check, SquareArrowOutUpRight} from "lucide-react";
import {isInteger} from "@/util/lib.ts";

const baseFormSchema = z.object({
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

const localConnectionFormSchema = z.object({
    connectionType: z.literal("local"),
});

const customConnectionFormSchema = z.object({
    connectionType: z.literal("host"),
    host: z.string()
        .min(1, "Database host is required")
        .refine(value => !value.includes(" "), {
            message: "Database host must not contain spaces",
        }),
    port: z.number()
        .min(1, "Database port is required")
        .max(65535, "Database port must be between 1 and 65535"),
    sslMode: z.enum(["disable", "require", "verify-ca", "verify-full"]),
    dbUsername: z.string()
        .min(1, "Database username is required")
        .refine(value => !value.includes(" "), {
            message: "Database username must not contain spaces",
        }),
    password: z.string()
        .min(1, "Database password is required"),
});

const formSchema = z.discriminatedUnion(
    "connectionType",
    [localConnectionFormSchema, customConnectionFormSchema],
).and(baseFormSchema);

const defaultValues: RepoPgInitDto = {
    version: 16,
    postgresPath: "",
    stopPostgres: true,
    connectionType: "host",
    postgresOsUser: "postgres",
    host: "localhost",
    port: 5432,
    sslMode: "disable",
    dbUsername: "postgres",
    password: "",
};

const PgSetup = () => {
    const {repoId: repoIdStr} = useParams<{ repoId: string }>();
    const repoId = parseInt(repoIdStr!);

    const pgImport = useNotifiableMutation({
        mutationKey: ["pg-import"],
        mutationFn: importPg,
        messages: {
            pending: "Starting PostgreSQL import",
            success: "PostgreSQL import started successfully",
        },
        invalidates: ["repo-list", "repo"],
    });

    const pgForm = useAppForm({
        defaultValues: defaultValues,
        resolver: zodResolver(formSchema),
    });

    const onSubmit = useCallback(async (repoPgInitDto: RepoPgInitDto) => {
        console.log(repoPgInitDto);
        await pgImport.mutateAsync({repoId, repoPgInitDto});
    }, [pgImport, repoId]);

    const hostConnectionSelected = pgForm.watch("connectionType") === "host";

    if (!isInteger(repoIdStr)) {
        return <Navigate to={"/error"} state={{message: "The repository ID in the URL is invalid."}}/>;
    }

    const repoInitSuccess = pgImport.isSuccess;
    const repoInitPending = pgImport.isPending;

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
                            <BreadcrumbPage>Import Data</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <h1 className={"mb-10"}>PostgreSQL Data Import</h1>

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
                        name="connectionType"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Connection Settings</FormLabel>
                                <Select
                                    disabled={repoInitPending || repoInitSuccess}
                                    defaultValue={field.value}
                                    onValueChange={field.onChange}>
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="How to connect to PostgreSQL?"/>
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="local">Use Local Connection</SelectItem>
                                        <SelectItem value="host">Use Custom Connection Configuration</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormDescription>
                                    <span className={"block"}>
                                        Local connection uses the PostgreSQL operating system user to connect to the
                                        database using <code className={"font-bold"}>trust</code> or <code
                                        className={"font-bold"}>peer</code> based auth.<br/>

                                    </span>

                                    <span className={"block mt-2"}>
                                        In case of a custom connection, you can provide a custom connection configuration.
                                    </span>

                                    <span className={"block mt-2"}>
                                       For either case, the given PostgreSQL user must have <code
                                        className={"font-bold"}>superuser</code> privileges.
                                    </span>
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

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
                                    className={"font-bold"}>postgres</code>
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    {hostConnectionSelected && (
                        <>
                            <div className={"flex gap-4"}>
                                <FormField
                                    control={pgForm.control}
                                    name="host"
                                    render={({field}) => (
                                        <FormItem className={"flex-1"}>
                                            <FormLabel>Host</FormLabel>
                                            <FormControl>
                                                <Input {...field}
                                                       disabled={true}
                                                       readOnly={true}
                                                       spellCheck="false"
                                                       placeholder="localhost"/>
                                            </FormControl>
                                            <FormDescription>
                                                PostgreSQL server hostname, currently only <code
                                                className={"font-bold"}>localhost</code> is
                                                supported
                                            </FormDescription>
                                            <FormMessage/>
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={pgForm.control}
                                    name="port"
                                    render={({field}) => (
                                        <FormItem className={"w-40"}>
                                            <FormLabel>Port</FormLabel>
                                            <FormControl>
                                                <Input {...field}
                                                       disabled={repoInitPending || repoInitSuccess}
                                                       type="number"
                                                       placeholder="5432"
                                                       onChange={(event) => {
                                                           const val = parseInt(event.target.value);
                                                           field.onChange(val);
                                                       }}/>
                                            </FormControl>
                                            <FormDescription>
                                                PostgreSQL server port
                                            </FormDescription>
                                            <FormMessage/>
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <FormField
                                control={pgForm.control}
                                name="sslMode"
                                render={({field}) => (
                                    <FormItem>
                                        <FormLabel>Postgres SSL Mode</FormLabel>
                                        <Select
                                            disabled={repoInitPending || repoInitSuccess}
                                            defaultValue={String(field.value)}
                                            onValueChange={(value: string) => {
                                                field.onChange(value);
                                            }}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select Postgres SSL Mode"/>
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="disable">disable</SelectItem>
                                                <SelectItem value="verify-ca">verify-ca</SelectItem>
                                                <SelectItem value="verify-full">verify-full</SelectItem>
                                                <SelectItem value="require">require</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormDescription>
                                            Select the SSL Mode for connecting to PostgreSQL.
                                            Use the <code
                                            className={"font-bold"}>disable</code> mode if you&#39;re unsure.<br/>
                                            Learn more about it in the&nbsp;
                                            <a
                                                target={"_blank"}
                                                rel={"noreferrer"}
                                                href={"https://www.postgresql.org/docs/current/libpq-ssl.html"}
                                                className={"text-blue-600"}>
                                                Postgres Documentation
                                                <SquareArrowOutUpRight className={"ms-0.5 inline relative top-[-1.5px]"}
                                                                       size={13}/>
                                            </a>
                                        </FormDescription>
                                        <FormMessage/>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={pgForm.control}
                                name="dbUsername"
                                render={({field}) => (
                                    <FormItem>
                                        <FormLabel>Username</FormLabel>
                                        <FormControl>
                                            <Input {...field}
                                                   disabled={repoInitPending || repoInitSuccess}
                                                   spellCheck="false"
                                                   placeholder="postgres"/>
                                        </FormControl>
                                        <FormDescription>
                                            PostgreSQL user name&nbsp;
                                            <i>
                                                (This MUST be a <code
                                                className={"font-bold"}>superuser</code>)
                                            </i>
                                        </FormDescription>
                                        <FormMessage/>
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={pgForm.control}
                                name="password"
                                render={({field}) => (
                                    <FormItem>
                                        <FormLabel>Password</FormLabel>
                                        <FormControl>
                                            <Input {...field}
                                                   disabled={repoInitPending || repoInitSuccess}
                                                   placeholder="••••"
                                                   type="password"/>
                                        </FormControl>
                                        <FormDescription>
                                            PostgreSQL user password <i>(This will NOT be saved)</i>
                                        </FormDescription>
                                        <FormMessage/>
                                    </FormItem>
                                )}
                            />
                        </>
                    )}

                    <div className={"flex gap-4"}>
                        <Button
                            type="submit"
                            variant={repoInitSuccess ? "success" : "default"}
                            disabled={repoInitPending || repoInitSuccess}>

                            <Spinner isLoading={repoInitPending}/>
                            {repoInitSuccess && <Check/>}
                            {repoInitSuccess ? "Import Started" : "Import Postgres Data"}
                        </Button>

                        {repoInitSuccess && (
                            <Link to={`/repo/${repoId}`}>
                                <Button>
                                    Go to Repo <ArrowRight/>
                                </Button>
                            </Link>
                        )}
                    </div>
                </form>
            </Form>

        </div>
    );
};

export default PgSetup;
