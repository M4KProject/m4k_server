go run ./init-rules init
go run ./pocketbase serve


# Création de l'utilisateur superuser si nécessaire
# go run ./pocketbase superuser create $PB_ADMIN_EMAIL $PB_ADMIN_PASSWORD

# GOOS=darwin GOARCH=amd64 go build -o pocketbase_mac ./pocketbase
# GOOS=linux GOARCH=amd64 go build -o pocketbase ./pocketbase
# GOOS=windows GOARCH=amd64 go build -o pocketbase.exe ./pocketbase