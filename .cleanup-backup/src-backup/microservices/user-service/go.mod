module user-service

go 1.25.0

replace github.com/jobfirst/jobfirst-core => ../../pkg/jobfirst-core

require (
	github.com/gin-gonic/gin v1.10.1
	github.com/hashicorp/consul/api v1.32.1
	github.com/jobfirst/jobfirst-core v0.0.0-00010101000000-000000000000
	gorm.io/gorm v1.25.5
)
