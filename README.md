# GoMongoServer
server.exe file requires only MongoDB to be installed on system

# Pre-requisites (Windows)
Make sure that you have Golang and its environment variable setup
MongoDB set up with environment variable in its place

# Pre-requisites (Mac or Linux)
Golang & MongoDB must be installed

# Running Code
Before running the code write following commands in VS Code Terminal
go get gopkg.in/mgo.v2
go get gopkg.in/mgo.v2/bson

# Building Project
go build server.go

# Running project via terminal
go run server.go

# Installing project globally
go install
