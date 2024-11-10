export interface PgLocalInitDto {
    postgresPath: string;
    version: number;
    stopPostgres: boolean;
    postgresOsUser: string;
}
