export interface RepoPgInitDto {
    postgresPath: string;
    version: number;
    stopPostgres: boolean;
    connectionType: string;
    postgresOsUser: string;
    host?: string;
    port?: number;
    dbUsername?: string;
    password?: string;
}
