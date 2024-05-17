package schema

import (
	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

var migrations = []darwin.Migration{{
	Version: 1,

	Description: "create table users",
	Script: `
	
	CREATE TABLE users(
		user_id UUID,
		name TEXT,
		email TEXT UNIQUE,
		roles TEXT[],
		password_hash TEXT,
		date_created TIMESTAMP,
		date_updated TIMESTAMP,

		PRIMARY KEY (user_id)
	);`,
}, {
	Version:     2,
	Description: "create table products",
	Script: `
	
	CREATE TABLE products(
		product_id UUID,
		name TEXT,
		cost INT,
		quantity INT,
		date_created TIMESTAMP,
		date_updated TIMESTAMP,
		PRIMARY KEY (product_id)
	);`,
},
	{Version: 3,
		Description: "create table sales",
		Script: `

		CREATE TABLE sales(
			sale_id UUID,
			product_id UUID,
			quantity INT,
			paid INT,
			date_created TIMESTAMP,
			date_updated TIMESTAMP,

			PRIMARY KEY (sale_id),
			FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE
		);`,
	}}

func Migrate(sb *sqlx.DB) error {
	driver := darwin.NewGenericDriver(sb.DB, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations)
	return d.Migrate()
}
