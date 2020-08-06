package database

// Health reports the status of the database connection.
func (d *Database) Health() error {
	return d.sqldb.Ping()
}
