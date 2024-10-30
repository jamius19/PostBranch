import {useMutation, UseMutationOptions, useQueryClient} from "@tanstack/react-query";
import {toast} from "react-toastify";
// @ts-expect-error Invalid Error
import {ToastContentProps} from "react-toastify/dist/types";
import {getRandomInt} from "@/util/lib.ts";

export const useNotifiableMutation = <TData, TError, TVariables>(
    options: UseMutationOptions<TData, TError, TVariables> & {
        messages?: {
            pending?: string;
            success?: string;
        },
        invalidates?: string[],
    }
) => {
    const queryClient = useQueryClient();

    const defaultMessages = {
        pending: "Running Operation",
        success: "Operation completed successfully!"
    };

    return useMutation({
        ...options,
        mutationFn: (variables) => {
            return toast.promise(
                new Promise((resolve, reject) => {
                    options.mutationFn!(variables)
                        .then(async (result) => {
                            await queryClient.invalidateQueries({
                                queryKey: options.invalidates || [],
                            });

                            setTimeout(() => resolve(result), getRandomInt(800, 1300));
                        })
                        .catch((error) => {
                            // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
                            setTimeout(() => reject(error), getRandomInt(300, 800));
                        });
                }),
                {
                    pending: options.messages?.pending || defaultMessages.pending,
                    success: options.messages?.success || defaultMessages.success,
                    error: {
                        render: (data: ToastContentProps<{ message: string, stack: string }>) => {
                            // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access,@typescript-eslint/no-unsafe-return
                            return data.data?.message || "An error occurred";
                        }
                    }
                }
            );
        }
    });
}