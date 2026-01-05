module github.com/szjason72/zervigo/services/core/tenant

go 1.21

require (
	github.com/gin-gonic/gin v1.10.1
	github.com/hashicorp/consul/api v1.32.1
	github.com/szjason72/zervigo/shared/core v0.0.0
	gorm.io/gorm v1.25.5
)

replace github.com/szjason72/zervigo/shared/core => ../../../shared/core

require (
	github.com/lib/pq v1.10.9
	gorm.io/driver/postgres v1.5.4
)

