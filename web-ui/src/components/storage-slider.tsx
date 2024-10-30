import {ChangeEvent, FocusEvent, KeyboardEvent, useCallback, useState} from 'react';
import {Slider} from "@/components/ui/slider";
import {Input} from "@/components/ui/input";
import {useController, UseControllerProps} from "react-hook-form";

const MIN_VALUE = 256; // 256 MB

const STEPS = [
    256, 512, 768, 1024, 1536, 2048, 3072, 4096, 5120, 6144, 7168, 8192, 9216, 10240, // Up to 10 GB
    12288, 14336, 16384, 20480, 24576, 28672, 32768, 40960, 49152, 57344, 65536,
    73728, 81920, 90112, 98304, // 10 GB to 96 GB in finer steps
    131072, 262144, 524288, 786432, 1048576, 1572864, 2097152 // Larger steps after 100 GB
];

const UNIT_MULTIPLIERS = new Map<string, number>([
    ['M', 1,],
    ['G', 1024,],
    ['T', 1024 * 1024]
]);

const regex = /^(\d+(?:\.\d+)?) ?([MGT])B?$/i;

interface StorageFormValues {
    sizeInMb?: number;
}

const StorageSlider = (props: UseControllerProps<StorageFormValues>) => {
    const {field, fieldState, formState} = useController<StorageFormValues>(props);

    const [sizeInMb, setSizeInMb] = useState(field.value || MIN_VALUE);
    const [inputValue, setInputValue] = useState(formatValue(field.value || MIN_VALUE, true));

    const getClosestStep = useCallback((val: number) => {
        return STEPS.reduce((prev, curr) =>
            Math.abs(curr - val) < Math.abs(prev - val) ? curr : prev
        )
    }, []);

    const handleSliderChange = useCallback((newValue: number[]) => {
        const stepIndex = Math.round((newValue[0] / 100) * (STEPS.length - 1));
        const value = STEPS[stepIndex];

        setSizeInMb(value);
        setInputValue(formatValue(value, true));
        field.onChange(value);
    }, [field]);

    const getSliderValue = useCallback(() => {
        const index = STEPS.indexOf(sizeInMb);
        if (index === -1) {
            const closestStep = getClosestStep(sizeInMb);
            const closestStepIndex = STEPS.indexOf(closestStep);
            return (closestStepIndex / (STEPS.length - 1)) * 100;
        }

        return (index / (STEPS.length - 1)) * 100;
    }, [getClosestStep, sizeInMb]);

    const handleInputChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
        const val = event.target.value;

        const matches = val.trim().match(regex);
        if (matches) {
            // number will contain the numeric value
            // unit will be 'M', 'G', or 'T'
            const [_, number, unit] = matches;

            const unitMultiplier = UNIT_MULTIPLIERS.get(unit.toUpperCase())!;
            const value = Math.ceil(Number(number) * unitMultiplier);

            setSizeInMb(value);
            field.onChange(value);
        } else {
            field.onChange(val);
        }

        setInputValue(val);
    }, [field]);

    const updateInputValue = useCallback((val: string) => {
        const matches = val.trim().match(regex);
        if (matches) {
            // number will contain the numeric value
            // unit will be 'M', 'G', or 'T'
            const [_, number, unit] = matches;

            const unitMultiplier = UNIT_MULTIPLIERS.get(unit.toUpperCase())!;
            const value = Number(number) * unitMultiplier;
            setInputValue(formatValue(value, true));
        }
    }, []);

    const onKeyDown = useCallback((event: KeyboardEvent<HTMLInputElement>) => {
        const target = event.target as HTMLInputElement;

        if (event.key === "Enter") {
            const val = target.value;
            const matches = val.trim().match(regex);
            if (matches) {
                updateInputValue(val);
            }

            target.blur();
        }
    }, [updateInputValue]);

    const onBlur = useCallback((event: FocusEvent<HTMLInputElement>) => {
        const val = (event.target as HTMLInputElement).value;
        updateInputValue(val);

    }, [updateInputValue]);

    return (
        <div className="w-full space-y-4">
            <div className="flex flex-col sm:flex-row items-center space-y-4 sm:space-y-0 sm:space-x-4">
                <Slider
                    min={0}
                    max={100}
                    step={1}
                    value={[getSliderValue()]}
                    onValueChange={handleSliderChange}
                    disabled={field.disabled}
                    className="w-full"/>

                <Input
                    ref={field.ref}
                    type="text"
                    value={inputValue}
                    onChange={handleInputChange}
                    onKeyDown={onKeyDown}
                    disabled={field.disabled}
                    onBlur={onBlur}
                    onFocus={(e) => e.target.select()}
                    className="w-24 storage-input"/>
            </div>
            <p className="text-sm text-gray-500 w-full sm:w-[800px]">
                Selected value: {formatValue(sizeInMb)}
            </p>
            {fieldState.error && (
                <p className="text-sm text-red-500">
                    {fieldState.error.message}
                </p>
            )}
        </div>
    )
}

const formatValue = (megabytes: number, shortUnitForm = false) => {
    let val: number;
    let unitSuffix: string;

    if (megabytes < 1024) {
        val = megabytes;

        if (shortUnitForm) {
            unitSuffix = "MB";
        } else {
            unitSuffix = val === 1 ? 'Megabyte' : 'Megabytes';
        }

        return `${megabytes} ${unitSuffix}`;
    } else if (megabytes < 1024 * 1024) {
        val = parseFloat((megabytes / 1024).toFixed(2));

        if (shortUnitForm) {
            unitSuffix = "GB";
        } else {
            unitSuffix = val === 1 ? 'Gigabyte' : 'Gigabytes';
        }

        return `${(megabytes / 1024).toFixed(2)} ${unitSuffix}`;
    } else {
        val = parseFloat((megabytes / (1024 * 1024)).toFixed(2));

        if (shortUnitForm) {
            unitSuffix = "TB";
        } else {
            unitSuffix = val === 1 ? 'Terabyte' : 'Terabytes';
        }

        return `${(megabytes / (1024 * 1024)).toFixed(2)} ${unitSuffix}`;
    }
};
export default StorageSlider;