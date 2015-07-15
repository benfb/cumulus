# cumulus
tool for injecting CoreOS cloud-configs into AWS Cloud Formation templates

### Download
`go get github.com/benfb/cumulus`

### Install
`go install github.com/benfb/cumulus`

### Example
This will format a cloud-config.yml into JSON format and inject it into the cloud-formation.json template, replacing whatever is between lines 146 and 267.

`cumulus inject --format cloud-config.yaml cloud-formation.json 146 267`
