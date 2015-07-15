# cumulus
tool for injecting CoreOS cloud-configs into AWS Cloud Formation templates

### Download
`go get github.com/benfb/cumulus`

### Install
`go install github.com/benfb/cumulus`

### Example
This will format a cloud-config.yml into JSON format and inject it into the cloud-formation.json template, replacing whatever is between lines 146 and 267.

`cumulus inject --format cloud-config.yaml cloud-formation.json 146 267`

Alternatively, you can just format the cloud-config and write it to a file.

`cumulus format cloud-config.yaml --out cloud-formation.json`

Or you can output to STDOUT and pipe the output somewhere else.

`cumulus format cloud-config.yaml | cat -n`
