import {BACKEND_API_URL} from "@/util/constants.ts";
import axios from "axios";
import {ResponseDto} from "@/@types/response-dto.ts";

const AxiosInstance = axios.create({
    baseURL: BACKEND_API_URL,
    headers: {
        'Content-Type': 'application/json'
    }
});

AxiosInstance.interceptors.response.use(
    // eslint-disable-next-line @typescript-eslint/no-unsafe-return
    response => response.data,
    error => {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access,@typescript-eslint/no-unsafe-assignment
        const data: ResponseDto<unknown> = error.response?.data;
        throw new Error(data.errors?.join(" ") || "An error occurred");
    }
);

export default AxiosInstance;
