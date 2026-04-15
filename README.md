# slurm-client-go

A command line tool for Slurm in CMA-HPC2023 by CEMC/CMA.

## Features

- Query active jobs
- Show partition information.
- (experiment) Filter long time jobs for operation users.

## Installing

Download the latest codes from GitHub.

Build the binary using:

```bash
go build -o bin/slclient main.go
```

Or use `Makefile` in Linux.

## Getting started

Query Slurm jobs:

```bash
slclient query
```

All jobs in queue will be shown:

```text
123456789 RUNNING operation   384 account1  2026-01-01 12:00:00 /path/to/account1/job1
987654321 RUNNING serial_op   1   account2  2026-01-01 12:30:00 /path/to/account2/job2
```

Use `slclient --help` to see more sub-commands.

## Commands

### `query`

Query active jobs in the Slurm queue.

```bash
slclient query
slclient query -u user1 -u user2
slclient query -p normal -s state:submit_time
slclient query -c "build.csh"
```

Flags:

- `-u, --user` — Filter by user (can be specified multiple times).
- `-p, --partition` — Filter by partition (can be specified multiple times).
- `-s, --sort-keys` — Sort keys separated by `:`, default is `state:submit_time`.
- `-c, --command-pattern` — Filter by command pattern.

### `info`

Show partition information.

```bash
slclient info
slclient info -s partition
```

Flags:

- `-s, --sort-keys` — Sort keys separated by `:`, default is `partition`.

### `detail`

Query jobs with detailed output (includes the full command line for each job).

```bash
slclient detail
slclient detail -u user1 -p normal
```

Flags:

- `-u, --user` — Filter by user (can be specified multiple times).
- `-p, --partition` — Filter by partition (can be specified multiple times).
- `-s, --sort-keys` — Sort keys separated by `:`, default is `state:submit_time`.
- `-c, --command-pattern` — Filter by command pattern.

### `watch`

Watch jobs until they finish. Checks every minute and prints a summary of job states.

```bash
slclient watch
slclient watch -j 1234567 -j 7654321
slclient watch -u user1 -p normal
```

Flags:

- `-u, --user` — Filter by user (can be specified multiple times).
- `-p, --partition` — Filter by partition (can be specified multiple times).
- `-j, --job` — Filter by job ID (can be specified multiple times).

### `filter` (experiment)

Filter long-running jobs for operation users.

```bash
slclient filter
```

### `category`

Show category list defined by command patterns.

```bash
slclient category
slclient category -d
```

Flags:

- `-d, --detail` — Show detailed information for each category.

## License

Copyright &copy; 2019-2026, developers at cemc-oper.

`slurm-client-go` is licensed under [MIT License](./LICENSE).