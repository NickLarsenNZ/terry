# Data Processing

Process schema versions for listed Terraform Providers

## Test and Build

Install [`Taskfile`](https://taskfile.dev)

```sh
task test  # Tests core library for all lambdas
task build # Builds all lambdas
```

For automatically running tests on code changes:

```sh
go test --watch
```

## Step Function flow

![Data Processing Step Function](../docs/images/stepfunction.png)
