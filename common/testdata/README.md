# Squeue test fixtures

These files are small, anonymized fixtures derived from real `squeue -o %all` output.
The original 52-column header, empty fields, state values, and duration formats are retained.
Usernames, accounts, job IDs, paths, node names, UIDs, timestamps, and resource sizes are synthetic.

- `squeue_sc1_anonymized.txt`: pending, running, and completing jobs; `MM:SS` and `HH:MM:SS` durations; an empty node list; a command containing spaces.
- `squeue_sc3_anonymized.txt`: pending, running, and completing jobs; `MM:SS`, `HH:MM:SS`, and day-qualified durations; `INVALID` time remaining.
- `squeue_array_anonymized.txt`: running and pending array tasks with distinct values in the duplicate `JOBID` and `PRIORITY` columns.

Keep these fixtures independent of the live cluster. When adding another real sample, anonymize it before adding it here and preserve the original column count and delimiters.
