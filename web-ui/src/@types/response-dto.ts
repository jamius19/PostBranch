export interface ResponseDto<T> {
    data: T;
    errors?: string[];
}