all: slurm_client
.PHONY: slurm_client

slurm_client:
	go build \
		-o bin/slclient \
		main.go