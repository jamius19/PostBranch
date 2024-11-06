export interface RepoPgInitDto {
    postgresPath: string;
    version: number;
    stopPostgres: boolean;
    connectionType: string;
    postgresOsUser: string;
    host?: string;
    port?: number;
    sslMode?: "verify-ca" | "verify-full" | "disable" | "require";
    dbUsername?: string;
    password?: string;
}
