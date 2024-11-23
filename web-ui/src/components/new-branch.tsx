import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import {Branch} from "@/@types/repo/repo-dto.ts";
import {Check, CopyPlus} from "lucide-react";
import {Button} from "@/components/ui/button.tsx";
import React, {useMemo, useState} from "react";
import {Input} from "@/components/ui/input.tsx";
import {useAppForm} from "@/lib/hooks/use-app-form.ts";
import {BranchInitDto} from "@/@types/repo/branch-init-dto.ts";
import {z} from "zod";
import {zodResolver} from "@hookform/resolvers/zod";
import {
    Form,
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage
} from "@/components/ui/form.tsx";
import BranchSelection from "@/components/branch-selection.tsx";
import {useNotifiableMutation} from "@/lib/hooks/use-notifiable-mutation.ts";
import {initBranch} from "@/service/branch-service.ts";
import Spinner from "@/components/spinner.tsx";

interface NewBranchProps {
    repoName: string;
    branches: Branch[];
    branchMap: Map<number, Branch>;
}

const defaultValues: BranchInitDto = {
    name: "",
    parentId: -1,
}

const NewBranch = ({repoName, branches, branchMap}: NewBranchProps) => {
    const [dialogOpen, setDialogOpen] = useState(false);

    const branchInit = useNotifiableMutation({
        mutationKey: ["branch-init"],
        mutationFn: (branchConfig: BranchInitDto) => initBranch(repoName, branchConfig),
        messages: {
            pending: "Creating Branch",
            success: "Branch created successfully",
        },
        invalidates: ["repo", repoName],
    });

    const formSchema = useMemo((): z.ZodType<BranchInitDto> => z.object({
        name: z.string()
            .min(1, "Branch name is required")
            .max(100, "Branch name must be less than 50 characters")
            .regex(/^[a-z][a-z0-9-]*[a-z0-9]$/, "Name must start with a letter, end with a letter or number, and can only contain letters, numbers, and hyphens")
            .refine(val => !branches.some(branch => branch.name === val),
                "Another branch with the same name already exists"),
        parentId: z.number()
            .min(1, "Parent Branch is required")
            .refine(val => {
                if (val === -1) {
                    return true;
                }

                const branch = branchMap.get(val)!;
                return branch.pgStatus !== "FAILED";
            }, "Branch Postgres can not be in FAILED state")
    }), [branches, branchMap]);

    const onSubmit = async (data: BranchInitDto) => {
        await branchInit.mutateAsync(data);
    };

    const branchForm = useAppForm({
        defaultValues,
        resolver: zodResolver(formSchema),
        mode: "onChange",
    });

    const dialogHandler = (open: boolean) => {
        setDialogOpen(open);

        if (!open) {
            branchForm.reset();
            branchInit.reset();
        }
    }

    return (
        <Dialog
            modal={true}
            onOpenChange={branchInit.isPending ? undefined : dialogHandler}
            open={branchInit.isPending ? true : dialogOpen}>

            <DialogTrigger asChild>
                <Button size={"sm"} variant={"outline"}>
                    <CopyPlus
                        size={13}
                        style={{width: 15, height: 15, marginRight: "-2px"}}
                        className={"relative top-[-0.5px]"}/>
                    New Branch
                </Button>
            </DialogTrigger>
            <DialogContent showClose={false} className={"max-w-[700px]"}>
                <DialogHeader>
                    <DialogTitle>Create new branch</DialogTitle>
                    <DialogDescription>
                        Choose a branch to create a new branch from
                    </DialogDescription>
                </DialogHeader>

                <Form {...branchForm}>
                    <form
                        onSubmit={branchForm.handleSubmit(onSubmit)}
                        onKeyDown={branchForm.disableSubmit}
                        className="w-2/3 space-y-8">

                        <FormField
                            control={branchForm.control}
                            name="name"
                            render={({field}) => (
                                <FormItem>
                                    <FormLabel>Branch Name</FormLabel>
                                    <FormControl>
                                        <Input {...field}
                                               disabled={branchInit.isPending || branchInit.isSuccess}
                                               placeholder="Enter Branch Name"
                                               spellCheck="false"
                                               className={"w-[450px]"}
                                               onChange={e => {
                                                   const val: string = e.target.value.toLowerCase().trim();
                                                   field.onChange(val);
                                               }}/>
                                    </FormControl>
                                    <FormDescription>
                                        Enter the name of the new branch
                                    </FormDescription>
                                    <FormMessage/>
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={branchForm.control}
                            name="parentId"
                            render={({field}) => (
                                <FormItem className="flex-1">
                                    <FormLabel>Select Parent Branch</FormLabel>
                                    <BranchSelection
                                        branches={branches}
                                        onBranchSelect={(value) => field.onChange(Number(value))}/>
                                    <FormDescription>
                                        Select the branch from which the new branch will be created. <br/>
                                        Only open branches are shown here.
                                    </FormDescription>
                                    <FormMessage/>
                                </FormItem>
                            )}
                        />

                        <div className={"flex gap-4"}>
                            <Button
                                type="submit"
                                variant={branchInit.isSuccess ? "success" : "default"}
                                disabled={branchInit.isPending || branchInit.isSuccess}>

                                <Spinner isLoading={branchInit.isPending} light/>
                                {branchInit.isSuccess && <Check/>}
                                {branchInit.isSuccess ? "Branch Created" : "Create Branch"}
                            </Button>

                            {(branchInit.isSuccess || branchInit.isError) && (
                                <Button onClick={() => dialogHandler(false)}>
                                    Close
                                </Button>
                            )}
                        </div>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
};

export default NewBranch;
