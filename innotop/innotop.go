package innotop

import "database/sql"

/*
+----------------------------+---------------------+------+---------------------+-------+
| Field                      | Type                | Null | Default             | Extra |
+----------------------------+---------------------+------+---------------------+-------+
| trx_id                     | varchar(18)         | NO   |                     |       |
| trx_state                  | varchar(13)         | NO   |                     |       |
| trx_started                | datetime            | NO   | 0000-00-00 00:00:00 |       |
| trx_requested_lock_id      | varchar(81)         | YES  | NULL                |       |
| trx_wait_started           | datetime            | YES  | NULL                |       |
| trx_weight                 | bigint(21) unsigned | NO   | 0                   |       |
| trx_mysql_thread_id        | bigint(21) unsigned | NO   | 0                   |       |
| trx_query                  | varchar(1024)       | YES  | NULL                |       |
| trx_operation_state        | varchar(64)         | YES  | NULL                |       |
| trx_tables_in_use          | bigint(21) unsigned | NO   | 0                   |       |
| trx_tables_locked          | bigint(21) unsigned | NO   | 0                   |       |
| trx_lock_structs           | bigint(21) unsigned | NO   | 0                   |       |
| trx_lock_memory_bytes      | bigint(21) unsigned | NO   | 0                   |       |
| trx_rows_locked            | bigint(21) unsigned | NO   | 0                   |       |
| trx_rows_modified          | bigint(21) unsigned | NO   | 0                   |       |
| trx_concurrency_tickets    | bigint(21) unsigned | NO   | 0                   |       |
| trx_isolation_level        | varchar(16)         | NO   |                     |       |
| trx_unique_checks          | int(1)              | NO   | 0                   |       |
| trx_foreign_key_checks     | int(1)              | NO   | 0                   |       |
| trx_last_foreign_key_error | varchar(256)        | YES  | NULL                |       |
| trx_adaptive_hash_latched  | int(1)              | NO   | 0                   |       |
| trx_adaptive_hash_timeout  | bigint(21) unsigned | NO   | 0                   |       |
| trx_is_read_only           | int(1)              | NO   | 0                   |       |
| trx_autocommit_non_locking | int(1)              | NO   | 0                   |       |
+----------------------------+---------------------+------+---------------------+-------+
*/

type InnoDBTransaction struct {
	ID      string `mysql:"trx_id"`
	State   string `mysql:"trx_state"`
	Started []byte // Currently we won't parse this due to time zone fun
	// it will be added later when we deal with timezones in the database tables?
	RequestedLockID      *sql.NullString `mysql:"trx_requested_lock_id"`
	WaitStarted          []byte          `mysql:"trx_wait_started"`
	Weight               uint64          `mysql:"trx_weight"`
	MySQLThreadID        uint64          `mysql:"trx_mysql_thread_id"`
	Query                *sql.NullString `mysql:"trx_query"`
	OperationState       *sql.NullString `mysql:"trx_operation_state"`
	TablesInUse          uint64          `mysql:"trx_tables_in_use"`
	TablesLocked         uint64          `mysql:"trx_tables_locked"`
	LockStructs          uint64          `mysql:"trx_lock_structs"`
	LockMemoryBytes      uint64          `mysql:"trx_lock_memory_bytes"`
	RowsLocked           uint64          `mysql:"trx_rows_locked"`
	RowsModified         uint64          `mysql:"trx_rows_modified"`
	ConcurrencyTickets   uint64          `mysql:"trx_concurrency_tickets"`
	IsolationLevel       string          `mysql:"trx_isolation_level"`
	UniqueChecks         int32           `mysql:"trx_unique_checks"`
	ForeignKeyChecks     int32           `mysql:"trx_foreign_key_checks"`
	LastForeignKeyError  *sql.NullString `mysql:"trx_last_foreign_key_error"`
	AdaptiveHashLatched  int32           `mysql:"trx_adaptive_hash_latched"`
	AdaptiveHashTimeout  uint64          `mysql:"trx_adaptive_hash_timeout"`
	IsReadOnly           int32           `mysql:"trx_is_read_only"`
	AutocommitNonLocking int32           `mysql:"trx_autocommit_non_locking"`
}
