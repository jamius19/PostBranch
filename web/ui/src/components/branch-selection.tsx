import * as React from "react"
import {Check, ChevronsUpDown} from "lucide-react"

import {cn} from "@/lib/utils"
import {Button} from "@/components/ui/button"
import {Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList,} from "@/components/ui/command";
import {Popover, PopoverContent, PopoverTrigger,} from "@/components/ui/popover";
import {Branch} from "@/@types/repo/repo-dto.ts";


interface BranchSelectionProps {
    branches: Branch[];
    onBranchSelect: (value: string) => void;
}

const BranchSelection = ({branches, onBranchSelect}: BranchSelectionProps) => {
    const [open, setOpen] = React.useState(false)
    const [value, setValue] = React.useState("")

    const branchData = React.useMemo(() =>
            branches
                .filter(branch => branch.status === "OPEN")
                .map(branch => ({
                    value: `${branch.id}`,
                    label: branch.name,
                })),
        [branches]);

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={open}
                    className={cn("w-[450px] justify-between border-input", open && "ring-1 ring-ring")}>

                    {value
                        ? branchData.find((framework) => framework.value === value)?.label
                        : "Select Branch..."}

                    <ChevronsUpDown className="opacity-50"/>
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-[450px] p-0">
                <Command>
                    <CommandInput placeholder="Search branch name..." className="h-9"/>
                    <CommandList>
                        <CommandEmpty>No branches found</CommandEmpty>
                        <CommandGroup>
                            {branchData.map((framework) => (
                                <CommandItem
                                    key={framework.value}
                                    value={framework.value}
                                    onSelect={(currentValue) => {
                                        setValue(currentValue);
                                        onBranchSelect(currentValue);
                                        setOpen(false);
                                    }}>

                                    {framework.label}

                                    <Check
                                        className={cn(
                                            "ml-auto",
                                            value === framework.value ? "opacity-100" : "opacity-0"
                                        )}/>
                                </CommandItem>
                            ))}
                        </CommandGroup>
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    )
}

export default BranchSelection;
