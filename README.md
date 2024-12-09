# Project Overview

## What is this?

This repository provides a foundational template for Golang-based backend applications utilizing gRPC communication, a PostgreSQL database, and a TypeScript frontend.

The template includes a basic user management system to streamline project setup and development.

## Why use this template ?

This template is designed to accelerate your project development by providing a solid starting point with the following features:

- Database Integration: Seamless integration with PostgreSQL, including code generation from database schema to ensure consistency.
- gRPC Communication: Efficient and reliable inter-service communication using the gRPC protocol.
- Proto File Generation: Automated generation of proto files to define service interfaces and data structures. Also ensure consistency and help upgrading the api.
- Database Migration: Simplified database schema management with migration tools.
- Testing Framework: Use of github action and go for testing the backend
- Cross-Platform Compatibility: The template leverages Docker to ensure consistent development environments across different operating systems, requiring only Golang, Docker, md5sum and bash to be installed.

By leveraging this template, you can focus on building your application's core functionality rather than spending time on boilerplate setup.

## Architecture

It s a monorepo

- backend: all the backend code in golang
    - backend/resources: resources files like database migration and html templates
    - backend/generated: generated files for backend
- common: Shared files between backend and frontend
    - common/proto: protofiles
    - common/scripts: all the script needed to generate and build the backend and frontend files
- frontend: Currently typescript frontend


# Installation and Usage

## requirements

- install bash
- install md5sum
- install docker
- install golang
- run in root directory to generate go.work file ([if you want to know why it s not commited](https://go.dev/ref/mod#go-work-file))
```
go work init
go work use backend
```
- replace `xxxyourappyyy` with the name of your app

## Environment variables

`VERIFICATION_EMAIL` : email use as from email for verification emails (default: `example@example.com`)

`VERIFICATION_TEMPLATE_FILE`: path to the verification template file use for the verification email (default: `verification_email.tmpl.html`)

`VERIFICATION_EMAIL_HOST`: http host for the verification page (default: `http://localhost:8080`)


`DBHOST`: Host of the psql server (default: `localhost`)

`DBPORT`: Port of the psql server (default: `5432`)

`DBPASSWD`: Password for the user on the psql server (default: ` `)

`DBUSER`: User for the psql server (default: `postgres`)

`DBNAME`: database name in the psql server (default: `postgres`)

`SMTPHOST`: Host of the smtp server(default: ` `)

`SMTPPORT`: Port of the smtp server (default: `587`)

`SMTPUSER`: User to authenticate with on the smtp server(default: ` `)

`SMTPPASSWD`: Password to authenticate with on the smtp server (default: ` `)



# Run server with local server

## run frontend 
```
bash common/scripts/run_local_frontend_server.sh
```

## run backend
```
bash common/scripts/generate_files.sh
cd backend
bash ../common/scripts/run_against_psql.sh  ./cmd/app/ --serve_bundle_url 'http://localhost:3535' --migration_file_path ./resources/database/migrations/ ../frontend/index.html  ../frontend/dist/ ./resources
```

# Questions

## Why are you using your own bash script?

I generally dislike most build systems. They can either:

- Interfere with your IDE's autocompletion capabilities.
- Be difficult to maintain.

In most cases, the time spent setting them up and maintaining them is not worthwhile. However, feel free to use or adapt a build system of your own preference.
