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
175294122 RUNNING operation   384 op_meso  2026-04-15 07:36:30 /g2/op_meso/OPER/ECFOUT/cma_meso_1km_v6_0_am/warm/06/model/fcst.job1
175294861 RUNNING serial_op   1   op_post  2026-04-15 07:50:02 /g2/op_post/ECFLOWOUT/cma_meso_1km_post_am/06/graph/meso_diag/area.shr_1km/plot_hour_003.job1
```

Use `slclient --help` to see more sub-commands.

## Commands

### `query`

Query active jobs in the Slurm queue.

```bash
slclient query
slclient query -u lijl -u chendh
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
slclient detail -u lijl -p normal
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
slclient watch -j 5831234 -j 5831591
slclient watch -u lijl -p normal
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