# Milestone1
Monolithic implementation of the trading server. What could go wrong?

## Installing
```shell
mkdir -p $GOPATH/src/github.com/distributeddesigns/milestone1
cd $GOPATH/src/github.com/distributeddesigns

# Note the change of M -> m; Fixes ref problem on case sensitive
# file systems (i.e. windows)
git clone https://github.com/DistributedDesigns/Milestone1.git milestone1

cd milestone1

go get

# creates a disposable temp binary; no need to `go install` then invoke
go run app.go ${workload file}
```
You can get workload files from the [docs repo][docs] or the [project website][project-website].

### Installing the linter
[metalinter][metalinter] will be run by CI. You can run the linter locally to check for problems early.

```shell
go get -u github.com/alecthomas/gometalinter

# Install all known linters
$GOPATH/bin/gometalinter --install

# Run the linter on all files the project
$GOPATH/bin/gometalinter --config=.gometalinterrc ./...

# Force gometalinter to pass before code is pushed to remote
chmod +x githooks/pre-push
ln -s $PWD/githooks/pre-push $PWD/.git/hooks/pre-push
```

### Never think about this again
Use a plugin to run the linter whenever you save a file in your IDE. [Here's a list.][linter-plugins] The only command you'll need to pass to the linter is `--config=.gometalinterrc`, which tells the linter to use the referenced config file.

## Validating logs
New logs for each run will be created in `./logs`. You can do partial validation for the schema using [logfile.xsd](./logfile.xsd) and `xmllint`.
```shell
xmllint --version # should be > 20624
xmllint --schema logfile.xsd --noout logs/yourlogfile.xml
```
Logfile troubleshooting is available on the [project website][logfile-faqs].

[docs]: https://github.com/distributeddesigns/docs
[project-website]: http://www.ece.uvic.ca/~seng462/ProjectWebSite/index.shtml
[logfile-faqs]: http://www.ece.uvic.ca/~seng462/ProjectWebSite/ExampleLog.html
[metalinter]: https://github.com/alecthomas/gometalinter
[linter-plugins]: https://github.com/alecthomas/gometalinter#editor-integration
