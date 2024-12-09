import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger
} from "@/components/ui/dialog.tsx";
import {Button} from "@/components/ui/button.tsx";
import {Trash2} from "lucide-react";
import React, {useState} from "react";

type DeleteRepoProps = {
    onDelete: () => void;
}

const DeleteRepo = (props: DeleteRepoProps) => {
    const [deleteModalOpen, setDeleteModalOpen] = useState(false);

    return (
        <div>
            <Dialog
                open={deleteModalOpen}
                onOpenChange={setDeleteModalOpen}>
                <DialogTrigger asChild>
                    <Button
                        className={"relative top-[-3px] ml-auto text-gray-400 px-2 py-2 hover:bg-red-600 hover:text-white hover:border-red-600 hover:shadow-md hover:shadow-red-500/40 transition-all duration-200"}
                        variant={"ghost"}>
                        <Trash2/>
                    </Button>
                </DialogTrigger>
                <DialogContent showClose={false} className={"max-w-[500px]"}>
                    <DialogHeader>
                        <DialogTitle>Delete Repository</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to delete the repository?<br/>
                            This action CANNOT be undone.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant={"outline"} onClick={() => setDeleteModalOpen(false)}>Cancel</Button>
                        <Button variant={"destructive"}
                                onClick={() => {
                                    setDeleteModalOpen(false);
                                    props.onDelete();
                                }}>
                            Delete
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default DeleteRepo;
