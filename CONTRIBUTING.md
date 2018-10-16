# Guidance On How To Contribute

There are two primary ways to help:
* Using the [issue tracker](https://github.com/SAP/quality-continuous-traceability-monitor/issues)
* Changing the code

## Using the [issue tracker](https://github.com/SAP/quality-continuous-traceability-monitor/issues)

Use the [issue tracker](https://github.com/SAP/quality-continuous-traceability-monitor/issues) to suggest feature requests, 
report bugs or ask questions. This is also a great way to connect with the developers of the project as well as others who
are interested in this tool.

## Changing the code

Use the [issue tracker](https://github.com/SAP/quality-continuous-traceability-monitor/issues) to find ways to contribute. 
Find a bug or a feature, mentioned in the issue that you will take care of.

Technically, you fork this repository, make changes in your own fork, and then submit a pull-request. 
All new code should have been thoroughly tested end-to-end in order to validate implemented features and the presence
or lack of defects.

### Working with forks
* [Configure this repository as a remote for your own fork](https://help.github.com/articles/configuring-a-remote-for-a-fork/)
* [Sync your fork with this repository](https://help.github.com/articles/syncing-a-fork/)   

before beginning to work on a new pull-request.

### Tests
All coding **must** come with automated [go unit tests](https://blog.alexellis.io/golang-writing-unit-tests/).

### Documentation
New or changed functionality needs to be documented, so it can be properly used.
Implementation of a functionality and its documentation shall happen within the same commit(s) and/or pull-request.

### Code Style

#### Formatting
You **must** run  `go fmt` on your changes to ensure proper [code formatting](https://golang.org/doc/effective_go.html#formatting) before you open a pull-request

#### Linting
You **must** run [golint](https://github.com/golang/lint) and solve all warnings before you open a pull-request

