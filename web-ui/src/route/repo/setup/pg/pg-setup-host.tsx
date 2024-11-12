import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {z} from "zod";
import {PgHostInitDto} from "@/@types/repo/pg/pg-host-init-dto.ts";
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
import Spinner from "@/components/Spinner.tsx";
import {ArrowRight, Check, Info, SquareArrowOutUpRight} from "lucide-react";
import Link from "@/components/Link.tsx";
import {formatValue} from "@/util/lib.ts";

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

const defaultValues: PgHostInitDto = {
    version: 16,
    postgresPath: "",
    host: "localhost",
    port: 5432,
    sslMode: "disable",
    dbUsername: "postgres",
    password: "",
};

const PgSetupHost = (): JSX.Element => {
    // const {repoId: repoIdStr} = useParams<{ repoId: string }>();
    // const repoId = parseInt(repoIdStr!);

    const pgValidate = useNotifiableMutation({
        mutationKey: ["pg-import"],
        mutationFn: (pgInit: PgHostInitDto) => validatePg(pgInit, "host"),
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

    const onSubmit = useCallback(async (pgInit: PgHostInitDto) => {
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

                    <hr className={"my-10 border-dotted"}/>

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
                                  state={{pgConfig: pgValidate.data.data!, adapter: "host"}}>
                                <Button>
                                    Configure Storage <ArrowRight/>
                                </Button>
                            </Link>
                        )}
                    </div>
                </form>
            </Form>
        </div>
    );
};

export default PgSetupHost;
