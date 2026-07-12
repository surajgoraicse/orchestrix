module github.com/surajgoraicse/orchestrix/services/coordinator

go 1.26.4

require (
	github.com/jackc/pgx/v5 v5.10.0
	github.com/surajgoraicse/orchestrix/api v0.0.0-00010101000000-000000000000
	github.com/surajgoraicse/orchestrix/libs/go-libs v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.82.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/surajgoraicse/orchestrix/api => ../../api

replace github.com/surajgoraicse/orchestrix/libs/go-libs => ../../libs/go-libs
