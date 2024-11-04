package dao

var PgSuperUserCheckQuery = `\du %s;`

var PgConfigFilePathsQuery = `SELECT DISTINCT(sourcefile) FROM pg_settings WHERE source = 'configuration file';`

var PgHbaFilePathsQuery = `SELECT DISTINCT(file_name) FROM pg_hba_file_rules;`

// TODO: Fix potential sql injection
var PgLocalReplicationCheckQuery = `SELECT CASE 
           WHEN EXISTS (
               SELECT 1
               FROM pg_hba_file_rules
               WHERE type = 'local'
                 AND 'replication' = ANY(database)
                 AND auth_method IN ('trust', 'peer')
                 AND ('%s' = ANY(user_name) OR 'all' = ANY(user_name))
           ) 
           THEN 'REPLICATION_ALLOWED'
           ELSE 'REPLICATION_NOT_ALLOWED'
       END AS replication_status;
`

// TODO: Fix potential sql injection
var PgHostReplicationCheckQuery = `SELECT CASE 
           WHEN EXISTS (
               SELECT 1
               FROM pg_hba_file_rules
               WHERE type = 'host'
                 AND 'replication' = ANY(database)
                 AND auth_method IN ('md5', 'scram-sha-256')
                 AND ('%s' = ANY(user_name) OR 'all' = ANY(user_name))
           ) 
           THEN 'REPLICATION_ALLOWED'
           ELSE 'REPLICATION_NOT_ALLOWED'
       END AS replication_status;
`
