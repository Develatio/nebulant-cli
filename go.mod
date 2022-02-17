module github.com/develatio/nebulant-cli

go 1.15

replace github.com/develatio/nebulant-cli => ./

require (
	github.com/aws/aws-sdk-go v1.34.24
	github.com/bhmj/jsonslice v1.1.1
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/go-playground/validator/v10 v10.10.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/gorilla/websocket v1.4.1
	github.com/joho/godotenv v1.3.0
	github.com/manifoldco/promptui v0.9.0
	github.com/povsister/scp v0.0.0-20210427074412-33febfd9f13e
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/mod v0.5.1
)
