package db

var migrationQueries = []string{
	identityMigration,
	serverConfigMigration,
	peerMigration,
	messageMigration,
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

var messageMigration = `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		peer_id TEXT,
		sender_id TEXT,
		text TEXT,
		timestamp INTEGER,
		read INTEGER DEFAULT 0
	);
`
