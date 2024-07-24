module github.com/MrAlias/otel-auto-demo/quota

go 1.22.4

require (
	github.com/MrAlias/otel-auto-demo/user v0.0.0-00010101000000-000000000000
	github.com/mattn/go-sqlite3 v1.14.22
)

replace github.com/MrAlias/otel-auto-demo/user => ../user/
