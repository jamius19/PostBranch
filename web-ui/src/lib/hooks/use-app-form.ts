import {FieldValues, useForm, UseFormProps, UseFormReturn} from "react-hook-form";
import {KeyboardEvent} from "react";

type DisableSubmitType = {
    disableSubmit: (e: KeyboardEvent<HTMLFormElement>) => void;
}

const disableSubmit = (e: KeyboardEvent<HTMLFormElement>) => {
    const target = e.target;
    if (e.key === "Enter" && target instanceof HTMLInputElement) {
        e.preventDefault();
    }
};


export const useAppForm =
    <T extends FieldValues>(props?: UseFormProps<T>): UseFormReturn<T> & DisableSubmitType => {
        const form = useForm<T>(props);

        return {
            ...form,
            disableSubmit
        }
    }
