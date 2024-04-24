# TODO

Todo is the backend of a minimalistic todo app. The guiding principle behind this project is to minimize
the usage of external dependencies and to try to rely on Go's standard library as much as possible, and to
explore the opportunities Go's `1.21` and `1.22` updates present for logging and routing.

Note:
The purpose of this repository is not to create a working product. It is mostly to experiment with different approaches to building go applications
and how different parts such as monitoring fit into the project.

# Requirements:
1. Go version 1.22 or greater
2. Docker and docker compose

### Setup
To run this project, Go version 1.22 or greater is required.
- Start by pulling the repository from Github:\
`git clone https://github.com/akalpaki/go-todo`
- Then, simply run `go mod tidy` to install dependencies.

And that's it!

### Usage
- `make run`: spins 
- `make test`: runs the application's test suite.

### Resources
Resources include most of the articles, repositories or in general resources I've used during research and exploration of the project:\
1. https://github.com/remisb/ : my mentor's github, with whom I bounce ideas back and forth and get inspiration
2. https://12factor.net/config : 12 Factor App configuration
3. https://www.youtube.com/@anthonygg_ : golang guru I get a lot of ideas from
4. https://github.com/urfave/negroni/tree/master : inspiration for access middleware solution
5. https://www.ardanlabs.com/ : I learned a lot from the Ultimate Go and Ultimate Go: Software Design with Kubernetes courses.
6. https://marcopeg.com/how-to-run-postgres-for-testing-in-docker/	
7. https://www.tonic.ai/blog/using-docker-to-manage-your-test-database
8. https://blog.postman.com/best-practices-for-api-error-handling/
9. https://www.calhoun.io/pitfalls-of-context-values-and-how-to-avoid-or-mitigate-them/
10. https://antonio-si.medium.com/a-few-tips-on-remote-debugging-golang-applications-running-in-an-m1-docker-container-68606326e83e
11. https://github.com/golang/vscode-go/wiki/debugging
12. https://github.com/go-delve/delve/tree/master/Documentation/cli
13. https://medium.com/metakratos-studio/debugging-golang-with-delve-6d5f0a1389aa
14. https://dev.to/bruc3mackenzi3/debugging-go-inside-docker-using-vscode-4f67
