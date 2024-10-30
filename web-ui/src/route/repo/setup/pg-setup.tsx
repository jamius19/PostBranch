import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb.tsx";
import {ChevronRightIcon} from "@radix-ui/react-icons";
import {Navigate, useParams} from "react-router-dom";
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
import {useCallback} from "react";
import {Checkbox} from "@/components/ui/checkbox.tsx";

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
});

const localConnectionFormSchema = z.object({
    customConnection: z.literal(false),
    postgresUser: z.string()
        .min(1, "PostgreSQL user is required")
        .refine(value => !value.includes(" "), {
            message: "PostgreSQL user must not contain spaces",
        }),
});

const customConnectionFormSchema = z.object({
    customConnection: z.literal(true),
    host: z.string()
        .min(1, "Database host is required")
        .refine(value => !value.includes(" "), {
            message: "Database host must not contain spaces",
        }),
    port: z.number()
        .min(1, "Database port is required")
        .max(65535, "Database port must be between 1 and 65535"),
    username: z.string()
        .min(1, "Database username is required")
        .refine(value => !value.includes(" "), {
            message: "Database username must not contain spaces",
        }),
    password: z.string()
        .min(1, "Database password is required"),
});

const formSchema = z.discriminatedUnion("customConnection", [localConnectionFormSchema, customConnectionFormSchema])
    .and(baseFormSchema);

const defaultValues: RepoPgInitDto = {
    version: 16,
    postgresPath: "",
    stopPostgres: true,
    customConnection: false,
    postgresUser: "postgres",
    host: "localhost",
    port: 5432,
    username: "postgres",
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
        invalidates: ["repo-list"],
    });

    const pgForm = useAppForm({
        defaultValues: defaultValues,
        resolver: zodResolver(formSchema),
    });

    const onSubmit = useCallback(async (repoPgInitDto: RepoPgInitDto) => {
        console.log(repoPgInitDto);
        await pgImport.mutateAsync({repoId, repoPgInitDto});
    }, [pgImport, repoId]);

    const customConnectionSelected = pgForm.watch("customConnection");

    if (!repoId) {
        return <Navigate to={"/error"}/>;
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
                        name="customConnection"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Connection Settings</FormLabel>
                                <Select
                                    defaultValue={String(field.value)}
                                    onValueChange={value => {
                                        const val: boolean = value.toLowerCase() === "true";
                                        field.onChange(val);
                                    }}>
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="How to connect to PostgreSQL?"/>
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="false">Use Local Connection</SelectItem>
                                        <SelectItem value="true">Use Custom Connection Configuration</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormDescription>
                                    <span className={"block"}>
                                        Local connection uses the PostgreSQL operating system user to connect to the
                                        database <i><b>without password</b></i><br/>
                                    This is recommended for typical PostgreSQL installations with PostBranch running
                                        on the same machine
                                    </span>

                                    <span className={"block mt-1.5"}>
                                        In case of a custom connection, you can provide a custom connection configuration
                                    </span>
                                </FormDescription>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    {customConnectionSelected ? (
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
                                                       readOnly={true}
                                                       spellCheck="false"
                                                       placeholder="localhost"/>
                                            </FormControl>
                                            <FormDescription>
                                                PostgreSQL server hostname, currently only <code>localhost</code> is
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
                                name="username"
                                render={({field}) => (
                                    <FormItem>
                                        <FormLabel>Username</FormLabel>
                                        <FormControl>
                                            <Input {...field}
                                                   spellCheck="false"
                                                   placeholder="postgres"/>
                                        </FormControl>
                                        <FormDescription>
                                            PostgreSQL user name
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
                                                   placeholder="••••"
                                                   type="password"/>
                                        </FormControl>
                                        <FormDescription>
                                            PostgreSQL user password
                                        </FormDescription>
                                        <FormMessage/>
                                    </FormItem>
                                )}
                            />
                        </>
                    ) : (
                        <FormField
                            control={pgForm.control}
                            name="postgresUser"
                            render={({field}) => (
                                <FormItem>
                                    <FormLabel>Postgres OS User</FormLabel>
                                    <FormControl>
                                        <Input {...field}
                                               spellCheck="false"
                                               placeholder={"postgres"}
                                               onChange={e => {
                                                   const val: string = e.target.value.trim();
                                                   field.onChange(val);
                                               }}/>
                                    </FormControl>
                                    <FormDescription>
                                        Local PostgreSQL operating system account&#39;s username<br/>
                                        If you&#39;re unsure, use the default value <i><b>postgres</b></i>
                                    </FormDescription>
                                    <FormMessage/>
                                </FormItem>
                            )}
                        />
                    )}

                    <Button type="submit">Import Postgres Data</Button>
                </form>
            </Form>

        </div>
    );
};

export default PgSetup;