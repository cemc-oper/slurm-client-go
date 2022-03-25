# slurm-client-go

A command line tool for Slurm using in CMA-PI HPC by CEMC.

## Features

- Query active jobs
- Show partition information.
- Filter long time jobs for operation users.

## Installing

Download the latest codes from Github.

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
5831234 RUNNING normal lijl    2019-01-29 00:37:49 /g6/lijl/BCC_CSMv3.v20190124/p25_2/build.csh
5831591 RUNNING normal lijl    2019-01-29 01:42:05 /g6/lijl/BCC_CSMv3.v20190124/p25_3/build.csh
5836542 RUNNING normal chendh  2019-01-29 08:41:45 /g8/JOB_TMP/chendh/ShCu/RUN_24/grapes.sbatch
5864521 RUNNING normal lijl    2019-01-31 01:10:34 /g6/lijl/BCC_CSMv3.v20190124/p25_5/build.csh
```

Use `slclient --help` to see more sub-commands.

## License

Copyright &copy; 2019-2022, Perilla Roc at cemc-oper.

`slurm-client-go` is licensed under [MIT License](./LICENSE).