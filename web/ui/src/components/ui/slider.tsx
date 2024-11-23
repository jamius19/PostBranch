import * as React from "react"
import * as SliderPrimitive from "@radix-ui/react-slider"

import {cn} from "@/lib/utils"
import {clsx} from "clsx";

const Slider =
    React.forwardRef<
        React.ElementRef<typeof SliderPrimitive.Root>,
        React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root>
    >(({className, ...props}, ref) => {
        return (
            <SliderPrimitive.Root
                ref={ref}
                className={cn(
                    "relative flex w-full touch-none select-none items-center",
                    className
                )}
                {...props}
            >
                <SliderPrimitive.Track
                    className={clsx("relative h-1.5 w-full grow overflow-hidden rounded-full", props.disabled ? "bg-primary/10" : "bg-primary/20")}>
                    <SliderPrimitive.Range
                        className={clsx("absolute h-full", props.disabled ? "bg-primary/40" : "bg-primary")}/>
                </SliderPrimitive.Track>
                <SliderPrimitive.Thumb
                    className="block h-4 w-4 rounded-full border border-primary/50 bg-background shadow transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50"/>
            </SliderPrimitive.Root>
        );
    })

Slider.displayName = SliderPrimitive.Root.displayName

export {Slider}
