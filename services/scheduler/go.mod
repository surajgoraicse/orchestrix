module github.com/surajgoraicse/orchestrix/services/scheduler

go 1.26.4

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo/v5 v5.3.0
	github.com/surajgoraicse/orchestrix/libs/go-libs v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)

replace github.com/surajgoraicse/orchestrix/libs/go-libs => ../../libs/go-libs
