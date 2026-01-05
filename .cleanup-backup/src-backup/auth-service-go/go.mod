module github.com/szjason72/zervigo/auth-service-go

go 1.25.0

replace github.com/jobfirst/jobfirst-core => ./jobfirst-core

require (
	github.com/jobfirst/jobfirst-core v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
)