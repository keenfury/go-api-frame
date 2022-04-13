# go-api-frame

Support library for [go-api-base](https://github.com/keenfury/go-api-base) so without the `go-api-base` library, this code base is worthless.

By combining this library and `go-api-base` you can build REST/GRPC *go* microservice.  If you are familiar with RoR scaffolding, well this is the same concept but with *go*.

### Go Versions
Built with Go 1.17+, verified 1.18

### Install
```
get clone github.com/keenfury/go-api-frame
cd go-api-frame
go mod tidy
go install .
```
This should put it in your `$GOPATH/bin` directory, make sure you have that in your `$PATH`.

Then you will need to add an `env var` to your system, called `FRAME_PATH` that points to `template` folder within this project.

### Usage
See [go-api-base](https://github.com/keenfury/go-api-base) cloning your new base code

Once installed, on your command line run this in your new project's root directory
```
go-api-frame
```
`go-api-frame` will ask a few starter questions:
- correct path to start the project
- which storage devices to create code? (you can choose one or all, if desired)
	- SQL
	- MongoDB
	- File
- If SQL is chosen then which engine?
	- Mysql
	- Postgres
	- Sqlite3
- If SQL is chosen, do you want to use `GORM`? (if you choose not to use `GORM`, `go-api-frame` will create sql statements for you, which I prefer, it is all about options)

Once this is completed then a `.frame` file is saved in the project's root directory.  With this in place then these questions will not be asked again.

The normal flow when starting this cli application within your project's root directory will ask questions to add a new endpoint based on SQL syntax.

How to add your *object* to the project.  The next menu asks this:
- Load File
- Paste Table Syntax
- Prompt for Table Syntax
- Blank Structure

Other than `Blank Structure` the options want a sql table syntax, the table name will become the name of the struct and the endpoint's grouping name.  For each group/endpoint it will create:
- GET - read by primary key
- GET - list all records
- POST - create new record
- PUT - update full record
- PATCH - update increment
- DELETE - delete by primary key

##### Boilerplate code
The general setup of boilerplate code follows the convention of [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).

Each layer is abstracted from each other by a Go interface.  Each layer can easily be tested without knowing *who* called it and *what* the downstream layer is.  See `manager_test.go` for some testing examples since most of the *business logic* would live in this file (layer).

Both a RESTful and GRPC server code is created for you.  See `cmd/rest` and `cmd/grpc`.

`go-api-base` has some hooks that this code base inserts for you, hence why this library and `go-api-base` go hand-in-hand.

### Notes
If `Blank Structure` is chosen, just know most of the boilerplate code will not be include but each file (layer) is created and, at least, should compile.  It is up to you then to add the logic for each layer.

If `Prompt for Table Syntax` is chosen, after files have been generated a sql structure is created for you and saved in .schema_prompt file will hold this, if you need to copy/paste for generating the table.

I've don't have all sql types represented and they may very from engine to engine on the effectiveness.  So this code generation is gives you an *as-is* end product, but it is *go* and you can change it how you need/want.

As of this writing, I have written but not tested multi-key functionality, so I can't say it works 100%.

You will be prompted as the end to do `go get update && go mod tidy`, make sure you do this.  As of this writing, if you specify *Postgres* as your engine you will need to do a `go-generate ./...` to get your mocks working correctly.  I think it all comes down that *go-api-base* has the library reference in it's *mod* files, *Postgres* is not and possibly the others as well.