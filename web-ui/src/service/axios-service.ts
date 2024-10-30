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
    response => response.data,
    error => {
        const data: ResponseDto<never> = error.response?.data;
        throw new Error(data.errors?.join(" ") || "An error occurred");
    }
);

export default AxiosInstance;