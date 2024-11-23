import * as React from "react"
import {cva, type VariantProps} from "class-variance-authority"

import {cn} from "@/lib/utils"

const badgeVariants = cva(
    "inline-flex items-center rounded-md border select-none font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
    {
        variants: {
            variant: {
                default:
                    "border-transparent bg-primary text-primary-foreground shadow hover:bg-primary/80",
                secondary:
                    "border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",
                destructive:
                    "border-transparent bg-destructive text-destructive-foreground shadow hover:bg-destructive/80",
                outline: "text-foreground",
                success: "border-transparent bg-lime-600 text-white hover:bg-lime-600/80",
                info: "border-blue-600 text-blue-600 hover:bg-blue-300/10",
            },
            size: {
                default: "px-2.5 py-0.5 text-xs",
                sm: "text-[0.6rem] px-2 py-0"
            }
        },
        defaultVariants: {
            variant: "default",
            size: "default",
        },
    }
)

export interface BadgeProps
    extends React.HTMLAttributes<HTMLDivElement>,
        VariantProps<typeof badgeVariants> {
}

function Badge({className, variant, size, ...props}: BadgeProps) {
    return (
        <div className={cn(badgeVariants({variant, size}), className)} {...props} />
    )
}

export {Badge, badgeVariants}
