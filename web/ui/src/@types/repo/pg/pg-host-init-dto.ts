export interface PgHostInitDto {
    postgresPath: string;
    version: number;
    host: string;
    port: number;
    sslMode: "verify-ca" | "verify-full" | "disable" | "require";
    dbUsername: string;
    password: string;
}
