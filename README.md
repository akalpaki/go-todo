# TODO

Todo is the backend of a minimalistic todo app. The guiding principle behind this project is to minimize
the usage of external dependencies and to try to rely on Go's standard library as much as possible, and to
explore the opportunities Go's `1.21` and `1.22` updates present for logging and routing.

Note: the `old` directory contains the initial hacked code I wrote when still figuring out the basics of the
project. It will be removed in a future version.

### Goals
1. Have a complete REST API that handles the registration and login of users.
The system will utilize JWT tokens for authorization
2. Have a complete REST API that handles the CRUD operations required for the todo application, based on user's authorization.
3. Have a solid configuration system that follows as much as possible the Twelve Factor App standard for application configuration. Configuration values should be orthogonal to each other and be provided as a set of modular values that do not depend on each other.
4. Do proper versioning.
5. Explore the patterns of middleware that are common in the Go community, such as the wrapper type around `http.Handler` which allows your handler functions to return errors that can be handled by your server in a proper manner.
6. Proper error handling. Error values returned to the consumer of the API should contain useful information to help debug issues, without exposing internal implementation. For the purposes of exploration, I will be using the Problem Details error handling style presented in [RFC-9457](https://datatracker.ietf.org/doc/html/rfc9457).
7. Provide a complete testing suite for the REST API. Completeness for the purposes of this project is not 100% code coverage, but a thorough testing of the API contracts which is done in such manner that internal implementation is not mocked. This way testing the contracts practically performs an end-to-end test.
8. Split the different domains handled by this server into separate autonomous microservices.
9. Use Docker to containerize the microservices and provide simple way to push them to production.
10. Explore CI/CD pipelines which can help streamline this process.
11. Use proper benchmarking, tracing and memory profiling to find optimization opportunities and discover how my coding style evolves as a concequence of the inefficiencies I uncover.

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
1. https://github.com/remisb/mat : standard golang project I use for inspiration
2. https://datatracker.ietf.org/doc/html/rfc9457 : the specification for `application/problem+json`
3. https://12factor.net/config : 12 Factor App configuration
4. https://www.youtube.com/@anthonygg_ : golang guru I get a lot of ideas from