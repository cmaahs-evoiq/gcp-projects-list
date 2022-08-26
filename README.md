# gcp-project-list

Quickly coded project list.

Cobbled together from all sorts of GCP golang reading...

It tries to determine your current GCP account/project from the
`~/.config/gcloud/configurations/config_default` file.  And tries to find
the `~/.config/gcloud/legacy_credentials/{account}/adc.json` json file to use
as the authentication to the REST API for GCP.

It will determine the Organization of your currently selected project and build
the folder/project list from there.

## build/run

```zsh
go mod tidy
go build
cp gcp-project-list /usr/local/bin/
gcp-project-list
```

## Limit to matching TOP level folder

```zsh
gcp-project-list <folder name>
```

## Add Folder and Project IDs

```zsh
gcp-project-list true
gcp-project-list <folder name> true
# and YES, if you have a folder with the name `true` at the first level under
# your organization, then you'll have some coding to do.
```
