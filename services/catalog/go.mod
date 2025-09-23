module wibusystem/services/catalog

go 1.25.1

replace (
	wibusystem/pkg/common => ../../pkg/common
	wibusystem/pkg/database => ../../pkg/database
	wibusystem/pkg/grpc => ../../pkg/grpc
	wibusystem/pkg/i18n => ../../pkg/i18n
)

require (
	github.com/gin-contrib/cors v1.7.6
	github.com/gin-gonic/gin v1.10.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/joho/godotenv v1.5.1
	wibusystem/pkg/common v0.0.0
	wibusystem/pkg/database v0.0.0
	wibusystem/pkg/i18n v0.0.0
)
