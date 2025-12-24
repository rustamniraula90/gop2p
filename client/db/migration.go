package db

var migrationQueries = []string{
	identityMigration,
	serverConfigMigration,
}

var identityMigration = `
	CREATE TABLE IF NOT EXISTS identity (
    	id TEXT PRIMARY KEY,
    	name TEXT UNIQUE NOT NULL
	);
`
var serverConfigMigration = `
	CREATE TABLE IF NOT EXISTS server_configs (
		address TEXT PRIMARY KEY,
		last_used INTEGER
	);
`

var peerMigration = `
	CREATE TABLE IF NOT EXISTS peers (
		id TEXT PRIMARY KEY,
		name TEXT ,
		ip TEXT ,
		port INTEGER,
		state INTEGER,
		last_used INTEGER
	)
`
