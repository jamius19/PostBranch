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
import React, {JSX, useCallback, useEffect, useRef} from "react";
import Spinner from "@/components/Spinner.tsx";
import {Check, SquareArrowOutUpRight} from "lucide-react";
import usePgAdapterState from "@/lib/hooks/use-pg-adapter-state.tsx";
import DbClusterInfo from "@/components/db-cluster-info.tsx";
import {scrollToElement} from "@/util/lib.ts";

const formSchema = z.object({
    version: z.number()
        .min(15, "Minimum supported Postgres version is 15")
        .max(17, "Maximum supported Postgres version is 17"),
    postgresPath: z.string()
        .min(1, "Postgres path is required")
        .refine(value => !value.includes(" "), {
            message: "Postgres path must not contain spaces",
        })
        .refine(value => value.startsWith("/") && !value.endsWith("/"), {
            message: "Postgres Path must start with '/' and not end with '/'",
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
    const pgValidate = useNotifiableMutation({
        mutationKey: ["pg-import"],
        mutationFn: (pgInit: PgHostInitDto) => validatePg(pgInit, "host"),
        messages: {
            pending: "Checking Postgres configuration",
            success: "Postgres configuration is valid",
        },
        invalidates: ["repo-list", "repo"],
    });

    const pgForm = useAppForm({
        defaultValues: defaultValues,
        resolver: zodResolver(formSchema),
    });

    const [Nav] = usePgAdapterState("host");

    const submitBtnRef = useRef<HTMLButtonElement>(null);

    const onSubmit = useCallback(async (pgInit: PgHostInitDto) => {
        await pgValidate.mutateAsync(pgInit);
    }, [pgValidate]);

    useEffect(() => {
        if (pgValidate.isSuccess) {
            scrollToElement(submitBtnRef.current!);
        }
    }, [pgValidate.isSuccess]);


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
                            <BreadcrumbPage>Configure Postgres</BreadcrumbPage>
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
            <h1 className={"mb-10"}>Postgres Connection</h1>

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
                                    Select the version of Postgres
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
                                <FormLabel>Local Postgres Installation Path</FormLabel>
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
                                    Absolute path to Postgres installation
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
                                                   spellCheck="false"
                                                   placeholder="localhost"/>
                                        </FormControl>
                                        <FormDescription>
                                            Postgres server hostname
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
                                            Postgres server port
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
                                        Select the SSL Mode for connecting to Postgres.
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
                                        Postgres user name&nbsp;
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
                                        Postgres user password <i>(This will NOT be saved)</i>
                                    </FormDescription>
                                    <FormMessage/>
                                </FormItem>
                            )}
                        />
                    </>

                    {repoInitSuccess && (
                        <DbClusterInfo clusterSizeInMb={pgValidate.data.data!.clusterSizeInMb}/>
                    )}

                    <div className={"flex gap-4"}>
                        <Button
                            ref={submitBtnRef}
                            type="submit"
                            variant={repoInitSuccess ? "success" : "default"}
                            disabled={repoInitPending || repoInitSuccess}>

                            <Spinner isLoading={repoInitPending} light/>
                            {repoInitSuccess && <Check/>}
                            {repoInitSuccess ? "Connected" : "Connect"}
                        </Button>

                        {repoInitSuccess && <Nav pgResponse={pgValidate.data.data!}/>}
                    </div>
                </form>
            </Form>
        </div>
    );
};

export default PgSetupHost;
