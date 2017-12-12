# Full Monte

_Full Monte_ is a collection of libraries to help implement [MCTS](https://en.wikipedia.org/wiki/Monte_Carlo_tree_search) for the general case! Proper documentation will come as features are completed and optimized.

Since MCTS is sometimes cited as an _embarassingly parallelisable_ algorithm, I thought it would be a good idea to plan ahead for the implementation of parallel MCTS. Not only this, but the plan is that fullmonte will make it very easy to implement MCTS over any domain necessary.

## Roadmap

- [x] interfaces to define abstract MCTS implementations
- [x] single-threaded MCTS
- [ ] optimization for general-case MCTS
- [ ] basic documentation
- [ ] concurrent multi-threaded MCTS
- [ ] a _worker-based_ concurrency model
  - [ ] leaf parallelisation
  - [ ] root parallelisation
  - [ ] tree parallelisation
- [ ] network worker support
- [ ] other base MCTS implementations (such as RAVE)

## Installing Dependencies and Running Tests

This project uses [glide](http://glide.sh/) for golang dependency management. This makes it very easy to fetch dependencies (as defined in glide.yaml) for the project. Glide is also very easy to install, check their website for more.

To download dependencies use the following:

```
glide install
```

This will download dependencies into a directory called `vendor`.

In order to run tests on the whole project without running tests for installed dependencies, use the following:

```
go test -v $(glide novendor)
```