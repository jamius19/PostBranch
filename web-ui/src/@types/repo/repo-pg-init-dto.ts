export interface RepoPgInitDto {
    postgresPath: string;
    version: number;
    stopPostgres: boolean;
    customConnection: boolean;
    postgresUser: string;
    host?: string;
    port?: number;
    username?: string;
    password?: string;
}