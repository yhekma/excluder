# Excluder

## Rationale

When using Time machine, you don't really have the option of excluding files/directories dynamically, you can only use static paths.
If you want to exclude all directories named `vendor` for instance, you are out of luck.

With this small tool you can define what to exclude in a config yaml (either by extension or dirname), run the tool, and it will set the xattributes on the matching files/dires so they will be excluded in the backups from then on.

# Usage

```
Usage of excluder:
  -config string
        yaml file with config (default "excluder.yaml")
  -verbose
        run in verbose mode
```

example of config:

```yaml
direxcludes:
  - .terraform
  - vendor

extextcludes:
  - .tmp
  - .swp
```