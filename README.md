# TODO

Todo is the backend of a minimalistic todo app. The guiding principle behind this project is to minimize
the usage of external dependencies and to try to rely on Go's standard library as much as possible, and to
explore the opportunities Go's `1.21` and `1.22` updates present for logging and routing.


### Goals
1. Create a complete backend for handling the needs of the todo app. In it's current version, this means a REST API for todo lists and users,
secured using JWT tokens. DONE
2. Integrate a PostgreSQL database for the persistence layer. Make the server and database work together using containers with docker-compose. DONE
3. Write testing suite.
4. Create CI/CD pipeline for deployment.
5. Use benchmarks, profiling and tracing to attempt to find any ways to improve the performance of the server.

### Current Progress:
REST API has been implemented. Now what is needed is:
- TLS
- API testing
- Containerization of server and integration with PostgreSQL using docker-compose.
- CI/CD implementation
- Versioning

### Setup
To run this project, Go version 1.22 or greater is required.
- Start by pulling the repository from Github:\
`git clone https://github.com/akalpaki/go-todo`
- Then, simply run `go mod tidy` to install dependencies.

And that's it!

### Usage
- `make run`: starts up the application
- `make test`: runs the application's test suite. WARNING: tests are ran using the `-race` flag. Expect longer testing times as a result of that.
- `make build`: builds the application

### Resources
Resources include most of the articles, repositories or in general resources I've used during research and exploration of the project:\
1. https://github.com/remisb/ : my mentor's github, with whom I bounce ideas back and forth and get inspiration
2. https://12factor.net/config : 12 Factor App configuration
3. https://www.youtube.com/@anthonygg_ : golang guru I get a lot of ideas from
4. https://github.com/urfave/negroni/tree/master : inspiration for access middleware solution